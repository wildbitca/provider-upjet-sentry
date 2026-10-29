package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

const examplesDir = "../examples"

var requiredParam = regexp.MustCompile(`^spec\.forProvider\.(\w+) is a required parameter$`)

// validate walks a manifest value against an OpenAPI v3 schema and reports
// unknown fields and type mismatches. It is an offline stand-in for server-side
// validation, which needs a cluster.
func validate(s *apiextv1.JSONSchemaProps, v any, path string, errs *[]string) {
	if s.XPreserveUnknownFields != nil && *s.XPreserveUnknownFields {
		return
	}
	switch x := v.(type) {
	case map[string]any:
		if s.Type != "" && s.Type != "object" {
			*errs = append(*errs, fmt.Sprintf("%s: object where %s expected", path, s.Type))
			return
		}
		for k, child := range x {
			if p, ok := s.Properties[k]; ok {
				validate(&p, child, path+"."+k, errs)
			} else if s.AdditionalProperties != nil && s.AdditionalProperties.Schema != nil {
				validate(s.AdditionalProperties.Schema, child, path+"."+k, errs)
			} else {
				*errs = append(*errs, fmt.Sprintf("%s.%s: unknown field", path, k))
			}
		}
	case []any:
		if s.Type != "array" || s.Items == nil || s.Items.Schema == nil {
			*errs = append(*errs, fmt.Sprintf("%s: array where %s expected", path, s.Type))
			return
		}
		for i, child := range x {
			validate(s.Items.Schema, child, fmt.Sprintf("%s[%d]", path, i), errs)
		}
	case string:
		if s.Type != "string" && !s.XIntOrString {
			*errs = append(*errs, fmt.Sprintf("%s: string where %s expected", path, s.Type))
		}
	case bool:
		if s.Type != "boolean" {
			*errs = append(*errs, fmt.Sprintf("%s: boolean where %s expected", path, s.Type))
		}
	case float64:
		if s.Type != "number" && s.Type != "integer" && !s.XIntOrString {
			*errs = append(*errs, fmt.Sprintf("%s: number where %s expected", path, s.Type))
		}
	}
}

func TestExamplesValidateAgainstCRDs(t *testing.T) {
	crds := loadCRDs(t)
	byGVK := map[string]*apiextv1.CustomResourceDefinition{}
	for _, c := range crds {
		byGVK[c.Spec.Group+"/"+c.Spec.Names.Kind] = c
	}

	var files []string
	err := filepath.WalkDir(examplesDir, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".yaml") {
			files = append(files, p)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(files)

	covered := map[string]bool{}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var obj map[string]any
		if err := yaml.Unmarshal(b, &obj); err != nil {
			t.Errorf("%s: %v", f, err)
			continue
		}
		apiVersion, _ := obj["apiVersion"].(string)
		kind, _ := obj["kind"].(string)
		group, version, _ := strings.Cut(apiVersion, "/")
		crd, ok := byGVK[group+"/"+kind]
		if !ok || version != "v1alpha1" {
			t.Errorf("%s: no CRD for %s %s", f, apiVersion, kind)
			continue
		}
		covered[group+"/"+kind] = true

		if string(crd.Spec.Scope) == "Namespaced" {
			if md, _ := obj["metadata"].(map[string]any); md["namespace"] == nil {
				t.Errorf("%s: namespaced resource without metadata.namespace", f)
			}
		}

		root := crd.Spec.Versions[0].Schema.OpenAPIV3Schema
		spec, _ := obj["spec"].(map[string]any)
		specSchema := root.Properties["spec"]
		var errs []string
		validate(&specSchema, spec, "spec", &errs)

		// Top-level "required parameter" rules, unless the manifest only observes.
		fp, _ := spec["forProvider"].(map[string]any)
		policies, _ := spec["managementPolicies"].([]any)
		observeOnly := len(policies) == 1 && policies[0] == "Observe"
		if !observeOnly {
			for _, rule := range specSchema.XValidations {
				m := requiredParam.FindStringSubmatch(rule.Message)
				if m == nil {
					continue
				}
				if _, ok := fp[m[1]]; ok {
					continue
				}
				if _, ok := fp[m[1]+"Ref"]; ok {
					continue
				}
				if _, ok := fp[m[1]+"Selector"]; ok {
					continue
				}
				errs = append(errs, fmt.Sprintf("spec.forProvider.%s is required", m[1]))
			}
		}
		for _, e := range errs {
			t.Errorf("%s: %s", f, e)
		}
	}

	// The namespaced tree has one example per group.
	for group, kind := range map[string]string{
		"organization": "Organization", "team": "Team", "project": "Project", "alert": "IssueAlert",
		"monitor": "CronMonitor", "integration": "IntegrationOpsgenie", "dashboard": "Dashboard",
	} {
		if key := group + ".upjet-sentry.m.upbound.io/" + kind; !covered[key] {
			t.Errorf("no namespaced example for %s", key)
		}
	}

	// One example per resource in the cluster tree.
	for name, want := range wantGroupKind {
		key := want[0] + ".upjet-sentry.upbound.io/" + want[1]
		if !covered[key] {
			t.Errorf("no example for %s (%s)", name, key)
		}
	}
}
