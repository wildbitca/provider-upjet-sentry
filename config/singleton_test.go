package config

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The two id-less singletons (ProjectOwnership, AllProjectsSpikeProtection)
// only work if the upjet build in use never writes an id into their synthetic
// state. These tests pin every piece of that assumption.

var idlessResources = map[string]string{
	"sentry_project_ownership":             "ProjectOwnership",
	"sentry_all_projects_spike_protection": "AllProjectsSpikeProtection",
}

// TestIDLessSchemasHaveNoIDAttribute pins the premise: the Terraform schema of
// both resources has no `id`, so upjet treats them as id-less.
func TestIDLessSchemasHaveNoIDAttribute(t *testing.T) {
	raw, err := os.ReadFile("schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		ProviderSchemas map[string]struct {
			ResourceSchemas map[string]struct {
				Block struct {
					Attributes map[string]json.RawMessage `json:"attributes"`
				} `json:"block"`
			} `json:"resource_schemas"`
		} `json:"provider_schemas"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	for name := range idlessResources {
		found := false
		for _, p := range doc.ProviderSchemas {
			if r, ok := p.ResourceSchemas[name]; ok {
				found = true
				if _, hasID := r.Block.Attributes["id"]; hasID {
					t.Errorf("%s now has an id attribute: its external-name strategy and the empty-id handling must be revisited", name)
				}
			}
		}
		if !found {
			t.Errorf("%s is missing from schema.json", name)
		}
	}
}

// TestIDLessCRDShape pins the generated CRD shape of both singletons in both
// API trees: the identity inputs are parameters, the id is observation only.
func TestIDLessCRDShape(t *testing.T) {
	crds := loadCRDs(t)
	wantParams := map[string][]string{
		"ProjectOwnership":           {"organization", "project", "raw"},
		"AllProjectsSpikeProtection": {"organization", "projects", "enabled"},
	}
	for _, kind := range idlessResources {
		seen := 0
		for _, crd := range crds {
			if crd.Spec.Names.Kind != kind {
				continue
			}
			seen++
			if len(crd.Spec.Versions) != 1 {
				t.Fatalf("%s: want one version, got %d", crd.Name, len(crd.Spec.Versions))
			}
			props := crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties
			forProvider := props["spec"].Properties["forProvider"].Properties
			atProvider := props["status"].Properties["atProvider"].Properties
			for _, p := range wantParams[kind] {
				if _, ok := forProvider[p]; !ok {
					t.Errorf("%s: spec.forProvider.%s is missing", crd.Name, p)
				}
			}
			if _, ok := forProvider["id"]; ok {
				t.Errorf("%s: spec.forProvider must not expose id", crd.Name)
			}
			if _, ok := atProvider["id"]; !ok {
				t.Errorf("%s: status.atProvider.id is missing", crd.Name)
			}
		}
		if seen != 2 {
			t.Errorf("%s: want a CRD in each of the two API trees, found %d", kind, seen)
		}
	}
}

// upjetSyntheticStateSource returns the files.go of the upjet build in use:
// the vendored copy when `make vendor` ran, otherwise the module in the cache.
func upjetSyntheticStateSource(t *testing.T) string {
	t.Helper()
	const rel = "pkg/terraform/files.go"
	vendored := filepath.Join("..", "vendor", "github.com", "crossplane", "upjet", "v2", rel)
	if b, err := os.ReadFile(vendored); err == nil {
		return string(b)
	}
	out, err := exec.CommandContext(t.Context(), "go", "list", "-m", "-f", "{{.Dir}}", "github.com/crossplane/upjet/v2").Output()
	dir := strings.TrimSpace(string(out))
	if err != nil || dir == "" {
		t.Fatalf("cannot locate the upjet module (go list: %v)", err)
	}
	b, err := os.ReadFile(filepath.Join(dir, rel))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestUpjetDoesNotWriteIDForIDLessResources pins how the upstream upjet build
// keeps the synthetic state valid for the id-less singletons: EnsureTFState
// writes `id` only when the Terraform schema has an `id` attribute (hasTFID, set
// from WithHasIDAttribute). Upstream does NOT skip an empty id for resources
// that do have one (crossplane/upjet#789), which matters for the
// IdentifierFromProvider resources; see docs/architecture/spikes.md.
func TestUpjetDoesNotWriteIDForIDLessResources(t *testing.T) {
	src := upjetSyntheticStateSource(t)
	if !strings.Contains(src, "if fp.hasTFID {") {
		t.Fatal("upjet no longer gates the synthetic state id on hasTFID: the id-less singletons (ProjectOwnership, AllProjectsSpikeProtection) would get an empty id; re-check the singleton external names")
	}
	if !strings.Contains(src, "WithHasIDAttribute") {
		t.Error("upjet no longer lets the store declare id-less resources (WithHasIDAttribute)")
	}
}
