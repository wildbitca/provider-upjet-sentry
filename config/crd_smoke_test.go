package config

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

const (
	crdDir = "../package/crds"
	// maxCRDBytes is the size budget per CRD: well under the 1 MiB etcd object
	// limit. CRDs above 256 KiB would also need server-side apply because of the
	// client-side-apply annotation limit.
	maxCRDBytes = 700 * 1024
)

func loadCRDs(t *testing.T) map[string]*apiextv1.CustomResourceDefinition {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(crdDir, "*.yaml"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no CRDs under %s (run make generate): %v", crdDir, err)
	}
	out := map[string]*apiextv1.CustomResourceDefinition{}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if len(b) > maxCRDBytes {
			t.Errorf("%s is %d bytes, over the %d byte budget", f, len(b), maxCRDBytes)
		}
		crd := &apiextv1.CustomResourceDefinition{}
		if err := yaml.UnmarshalStrict(b, crd); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		out[crd.Name] = crd
	}
	return out
}

// TestCRDsPerGroup is the per-group smoke test: every group has the expected
// Kinds in both API trees, with the right scope, one served and stored version
// and a non-empty structural schema.
func TestCRDsPerGroup(t *testing.T) {
	crds := loadCRDs(t)
	const (
		clusterRoot    = "upjet-sentry.upbound.io"
		namespacedRoot = "upjet-sentry.m.upbound.io"
	)

	for name, want := range wantGroupKind {
		group, kind := want[0], want[1]
		for _, tree := range []struct{ root, scope string }{
			{clusterRoot, "Cluster"},
			{namespacedRoot, "Namespaced"},
		} {
			var found *apiextv1.CustomResourceDefinition
			for _, c := range crds {
				if c.Spec.Group == group+"."+tree.root && c.Spec.Names.Kind == kind {
					found = c
				}
			}
			if found == nil {
				t.Errorf("%s: no CRD for %s.%s kind %s", name, group, tree.root, kind)
				continue
			}
			if string(found.Spec.Scope) != tree.scope {
				t.Errorf("%s: scope %s, want %s", found.Name, found.Spec.Scope, tree.scope)
			}
			if len(found.Spec.Versions) != 1 || !found.Spec.Versions[0].Served || !found.Spec.Versions[0].Storage {
				t.Errorf("%s: want exactly one served+storage version, got %+v", found.Name, found.Spec.Versions)
				continue
			}
			spec := found.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"]
			fp, ok := spec.Properties["forProvider"]
			if !ok || len(fp.Properties) == 0 {
				t.Errorf("%s: empty spec.forProvider", found.Name)
			}
		}
	}

	// 24 resources x 2 trees + ProviderConfig kinds, nothing else.
	var resources, other []string
	for n, c := range crds {
		if strings.HasPrefix(c.Spec.Group, "upjet-sentry.") {
			other = append(other, n)
		} else {
			resources = append(resources, n)
		}
	}
	sort.Strings(other)
	if len(resources) != 2*len(wantGroupKind) {
		t.Errorf("%d resource CRDs, want %d", len(resources), 2*len(wantGroupKind))
	}
	wantOther := []string{
		"clusterproviderconfigs.upjet-sentry.m.upbound.io",
		"providerconfigs.upjet-sentry.m.upbound.io",
		"providerconfigs.upjet-sentry.upbound.io",
		"providerconfigusages.upjet-sentry.m.upbound.io",
		"providerconfigusages.upjet-sentry.upbound.io",
	}
	if strings.Join(other, ",") != strings.Join(wantOther, ",") {
		t.Errorf("root-group CRDs = %v, want %v", other, wantOther)
	}
}

// TestNoPlainSecretsInCRDs makes sure the sensitive inputs became secret
// references and that no secret-bearing attribute leaked into a CRD as a plain
// field.
func TestNoPlainSecretsInCRDs(t *testing.T) {
	crds := loadCRDs(t)
	forbidden := regexp.MustCompile(`(?m)^\s+(dsnSecret|dsnPublic|dsnCsp|securityToken|integrationKey|privateKey|secretKey|appConnectPrivateKey|password):`)
	for name, c := range crds {
		b, err := yaml.Marshal(c)
		if err != nil {
			t.Fatal(err)
		}
		if m := forbidden.FindString(string(b)); m != "" {
			t.Errorf("%s exposes plain secret attribute %q", name, strings.TrimSpace(m))
		}
	}
	wantRefs := map[string][]string{
		"projectsymbolsources.project.upjet-sentry.upbound.io":       {"appConnectPrivateKeySecretRef", "passwordSecretRef", "privateKeySecretRef", "secretKeySecretRef"},
		"integrationopsgenies.integration.upjet-sentry.upbound.io":   {"integrationKeySecretRef"},
		"integrationpagerduties.integration.upjet-sentry.upbound.io": {"integrationKeySecretRef"},
		"projects.project.upjet-sentry.upbound.io":                   {"securityTokenSecretRef"},
	}
	for name, refs := range wantRefs {
		c, ok := crds[name]
		if !ok {
			t.Errorf("CRD %s not found", name)
			continue
		}
		b, _ := yaml.Marshal(c)
		for _, r := range refs {
			if !strings.Contains(string(b), r+":") {
				t.Errorf("%s is missing %s", name, r)
			}
		}
	}
}

// TestReferenceFieldsExistInCRDs asserts that representative references reach
// the generated API as <field>Ref / <field>Selector pairs, and that fields that
// must stay plain values did not grow one.
func TestReferenceFieldsExistInCRDs(t *testing.T) {
	crds := loadCRDs(t)
	forProvider := func(name string) map[string]apiextv1.JSONSchemaProps {
		c, ok := crds[name]
		if !ok {
			t.Fatalf("CRD %s not found", name)
		}
		return c.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"].Properties["forProvider"].Properties
	}
	for _, root := range []string{"upjet-sentry.upbound.io", "upjet-sentry.m.upbound.io"} {
		has := map[string][]string{
			"teams.team." + root:                          {"organizationRef", "organizationSelector"},
			"projects.project." + root:                    {"organizationRef", "teamsRefs", "teamsSelector"},
			"projectkeys.project." + root:                 {"organizationRef", "projectRef", "projectSelector"},
			"teammembers.team." + root:                    {"organizationRef", "teamRef", "memberIdRef"},
			"allprojectsspikeprotections.project." + root: {"projectsRefs", "projectsSelector"},
			"notificationactions.alert." + root:           {"projectsRefs", "projectsSelector"},
			"issuealerts.alert." + root:                   {"projectRef"},
			"integrationpagerduties.integration." + root:  {"organizationRef"},
			"dashboards.dashboard." + root:                {"organizationRef"},
		}
		for crd, fields := range has {
			props := forProvider(crd)
			for _, f := range fields {
				if _, ok := props[f]; !ok {
					t.Errorf("%s: missing %s", crd, f)
				}
			}
		}
		// owner.teamId is a nested reference on the monitors.
		for _, k := range []string{"cronmonitors", "metricmonitors", "uptimemonitors"} {
			owner := forProvider(k + ".monitor." + root)["owner"]
			for _, f := range []string{"teamIdRef", "teamIdSelector"} {
				if _, ok := owner.Properties[f]; !ok {
					t.Errorf("%s.owner: missing %s", k, f)
				}
			}
		}
		// Plain values by design.
		hasNot := map[string][]string{
			"integrationopsgenies.integration." + root:      {"teamRef", "teamSelector", "integrationIdRef"},
			"organizationcodemappings.organization." + root: {"projectIdRef", "projectIdSelector"},
			"organizations.organization." + root:            {"organizationRef"},
		}
		for crd, fields := range hasNot {
			props := forProvider(crd)
			for _, f := range fields {
				if _, ok := props[f]; ok {
					t.Errorf("%s: %s must not exist", crd, f)
				}
			}
		}
	}
}
