package config

import (
	"context"
	"encoding/json"
	"os"
	"sort"
	"strings"
	"testing"
)

// TestResourcesMatchTheSchema fails when the pinned provider schema gains or
// loses a resource, so a provider bump forces a conscious decision about group,
// Kind and external name instead of silently skipping the new resource.
func TestResourcesMatchTheSchema(t *testing.T) {
	b, err := os.ReadFile("schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var s struct {
		ProviderSchemas map[string]struct {
			ResourceSchemas map[string]json.RawMessage `json:"resource_schemas"`
		} `json:"provider_schemas"`
	}
	if err := json.Unmarshal(b, &s); err != nil {
		t.Fatal(err)
	}
	if len(s.ProviderSchemas) != 1 {
		t.Fatalf("want exactly one provider schema, got %d", len(s.ProviderSchemas))
	}
	var got []string
	for _, ps := range s.ProviderSchemas {
		for name := range ps.ResourceSchemas {
			got = append(got, name)
		}
	}
	sort.Strings(got)
	want := append([]string(nil), wantResources...)
	sort.Strings(want)
	if len(got) != len(want) {
		t.Fatalf("schema has %d resources %v, want %d %v", len(got), got, len(want), want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("schema resource %s differs from the contract (%s)", got[i], want[i])
		}
	}
}

// stateContract describes, per resource, the Terraform state `id` the upstream
// provider stores (upstream-provider.md section 2) and the external name derived
// from it. GetIDFn(externalName, parameters) must reproduce the state id and
// GetExternalNameFn(state) must recover the external name.
type stateContract struct {
	params map[string]any
	// stateID is the `id` attribute stored by the provider; empty for resources
	// that have no `id` attribute at all.
	stateID string
	// state holds extra attributes present in the stored state.
	state        map[string]any
	externalName string
	// importID is what GetIDFn must return when it differs from stateID
	// (id-less resources, where it is the import string).
	importID string
}

var contracts = map[string]stateContract{
	"sentry_organization":              {params: map[string]any{}, stateID: "acme", externalName: "acme"},
	"sentry_team":                      {params: map[string]any{"organization": "acme"}, stateID: "backend", externalName: "backend"},
	"sentry_project":                   {params: map[string]any{"organization": "acme"}, stateID: "web", externalName: "web"},
	"sentry_organization_code_mapping": {params: map[string]any{"organization": "acme"}, stateID: "31", externalName: "31"},
	"sentry_organization_repository":   {params: map[string]any{"organization": "acme"}, stateID: "88", externalName: "88"},
	"sentry_key":                       {params: map[string]any{"organization": "acme", "project": "web"}, stateID: "keyid", externalName: "keyid"},
	"sentry_project_symbol_source":     {params: map[string]any{"organization": "acme", "project": "web"}, stateID: "src1", externalName: "src1"},
	"sentry_issue_alert":               {params: map[string]any{"organization": "acme", "project": "web"}, stateID: "55", externalName: "55"},
	"sentry_alert":                     {params: map[string]any{"organization": "acme"}, stateID: "66", externalName: "66"},
	"sentry_notification_action":       {params: map[string]any{"organization": "acme"}, stateID: "77", externalName: "77"},
	"sentry_cron_monitor":              {params: map[string]any{"organization": "acme", "project": "web"}, stateID: "101", externalName: "101"},
	"sentry_metric_monitor":            {params: map[string]any{"organization": "acme", "project": "web"}, stateID: "102", externalName: "102"},
	"sentry_uptime_monitor":            {params: map[string]any{"organization": "acme", "project": "web"}, stateID: "103", externalName: "103"},
	"sentry_integration_opsgenie":      {params: map[string]any{"organization": "acme"}, stateID: "svc1", externalName: "svc1"},
	"sentry_integration_pagerduty":     {params: map[string]any{"organization": "acme"}, stateID: "svc2", externalName: "svc2"},

	"sentry_organization_member": {params: map[string]any{"organization": "acme"}, stateID: "acme/42", externalName: "42"},
	"sentry_metric_alert":        {params: map[string]any{"organization": "acme", "project": "web"}, stateID: "acme/web/7", externalName: "7"},
	"sentry_dashboard":           {params: map[string]any{"organization": "acme"}, stateID: "acme/9", externalName: "9"},

	"sentry_plugin": {
		params:  map[string]any{"organization": "acme", "project": "web", "plugin": "slack"},
		stateID: "slack", state: map[string]any{"plugin": "slack"}, externalName: "slack", importID: "slack",
	},
	"sentry_team_member": {
		params:  map[string]any{"organization": "acme", "team": "backend", "member_id": "11"},
		stateID: "acme/backend/11", state: map[string]any{"member_id": "11"}, externalName: "11", importID: "acme/backend/11",
	},
	"sentry_project_inbound_data_filter": {
		params:  map[string]any{"organization": "acme", "project": "web", "filter_id": "browser-extensions"},
		stateID: "acme/web/browser-extensions", state: map[string]any{"filter_id": "browser-extensions"}, externalName: "browser-extensions", importID: "acme/web/browser-extensions",
	},
	"sentry_project_spike_protection": {
		params:  map[string]any{"organization": "acme", "project": "web"},
		stateID: "acme/web", state: map[string]any{"organization": "acme", "project": "web"}, externalName: "web", importID: "acme/web",
	},
	"sentry_project_ownership": {
		params:  map[string]any{"organization": "acme", "project": "web"},
		stateID: "", state: map[string]any{"organization": "acme", "project": "web"}, externalName: "acme/web", importID: "acme/web",
	},
	"sentry_all_projects_spike_protection": {
		params:  map[string]any{"organization": "acme"},
		stateID: "", state: map[string]any{"organization": "acme"}, externalName: "acme", importID: "acme",
	},
}

func TestEveryResourceHasAStateContract(t *testing.T) {
	for _, name := range wantResources {
		if _, ok := contracts[name]; !ok {
			t.Errorf("no state contract for %s", name)
		}
		e, ok := ExternalNameConfigs[name]
		if !ok {
			t.Errorf("no external name config for %s", name)
			continue
		}
		if e.GetIDFn == nil || e.GetExternalNameFn == nil || e.SetIdentifierArgumentFn == nil {
			t.Errorf("%s: GetIDFn, GetExternalNameFn and SetIdentifierArgumentFn must all be set", name)
		}
	}
}

// TestExternalNameRoundTrip checks, for all 24 resources, that the synthetic
// state id Upjet writes equals the id the provider stores, and that the
// external name is recovered from a stored state.
func TestExternalNameRoundTrip(t *testing.T) {
	for name, c := range contracts {
		t.Run(name, func(t *testing.T) {
			e := ExternalNameConfigs[name]

			gotID, err := e.GetIDFn(context.Background(), c.externalName, c.params, nil)
			if err != nil {
				t.Fatalf("GetIDFn: %v", err)
			}
			want := c.stateID
			if c.importID != "" {
				want = c.importID
			}
			if gotID != want {
				t.Errorf("GetIDFn = %q, want %q", gotID, want)
			}

			state := map[string]any{}
			for k, v := range c.state {
				state[k] = v
			}
			if c.stateID != "" {
				state["id"] = c.stateID
			}
			gotEN, err := e.GetExternalNameFn(state)
			if err != nil {
				t.Fatalf("GetExternalNameFn: %v", err)
			}
			if gotEN != c.externalName {
				t.Errorf("GetExternalNameFn = %q, want %q", gotEN, c.externalName)
			}
		})
	}
}

// importStrings is the string `terraform import` must receive, per resource,
// as required by the importers of jianyuan/sentry v0.15.8 (internal/provider
// intresource.ImportState*Part* and sentry/importer.go). Resources not listed
// here import with the same string they store as the state id.
var importStrings = map[string]string{
	"sentry_team":                      "acme/backend",
	"sentry_project":                   "acme/web",
	"sentry_organization_code_mapping": "acme/31",
	"sentry_organization_repository":   "acme/github/12/88",
	"sentry_key":                       "acme/web/keyid",
	"sentry_project_symbol_source":     "acme/web/src1",
	"sentry_issue_alert":               "acme/web/55",
	"sentry_alert":                     "acme/66",
	"sentry_notification_action":       "acme/77",
	"sentry_cron_monitor":              "acme/101",
	"sentry_metric_monitor":            "acme/102",
	"sentry_uptime_monitor":            "acme/103",
	"sentry_integration_opsgenie":      "acme/5/svc1",
	"sentry_integration_pagerduty":     "acme/5/svc2",
	"sentry_plugin":                    "acme/web/slack",
}

// importParams adds the extra parameters some importers need on top of the
// contract parameters.
var importParams = map[string]map[string]any{
	"sentry_organization_repository": {"integration_type": "github", "integration_id": "12"},
	"sentry_integration_opsgenie":    {"integration_id": "5"},
	"sentry_integration_pagerduty":   {"integration_id": "5"},
}

// TestImportIDRoundTrip is the Observe-only/adoption contract: the string
// handed to `terraform import` is the composite the provider importer expects,
// while GetIDFn keeps returning the id the provider stores in state. It
// regresses the v0.1.0 bug where a bare id was imported and the importer
// rejected it.
func TestImportIDRoundTrip(t *testing.T) {
	ctx := context.Background()
	for name, c := range contracts {
		t.Run(name, func(t *testing.T) {
			e := ExternalNameConfigs[name]
			params := map[string]any{}
			for k, v := range c.params {
				params[k] = v
			}
			for k, v := range importParams[name] {
				params[k] = v
			}

			stateID, err := e.GetIDFn(ctx, c.externalName, params, nil)
			if err != nil {
				t.Fatalf("GetIDFn: %v", err)
			}
			if c.stateID != "" && !strings.HasSuffix(stateID, c.stateID) {
				t.Errorf("GetIDFn = %q, want the stored state id %q", stateID, c.stateID)
			}

			want, composite := importStrings[name]
			if !composite {
				if e.GetImportIDFn != nil {
					got, err := e.GetImportIDFn(ctx, c.externalName, params, nil)
					if err != nil || got != stateID {
						t.Errorf("GetImportIDFn = %q, %v; want the state id %q", got, err, stateID)
					}
				}
				return
			}
			if e.GetImportIDFn == nil {
				t.Fatalf("GetImportIDFn is nil, the importer needs %q", want)
			}
			got, err := e.GetImportIDFn(ctx, c.externalName, params, nil)
			if err != nil {
				t.Fatalf("GetImportIDFn: %v", err)
			}
			if got != want {
				t.Errorf("GetImportIDFn = %q, want %q", got, want)
			}
			// The importer stores the last path segment as the state id.
			if last := got[strings.LastIndex(got, "/")+1:]; last != c.stateID {
				t.Errorf("last import segment %q differs from the stored state id %q", last, c.stateID)
			}
			if stateID != c.stateID {
				t.Errorf("GetIDFn = %q, must stay the bare state id %q", stateID, c.stateID)
			}

			empty, err := e.GetImportIDFn(ctx, "", params, nil)
			if err != nil || empty != "" {
				t.Errorf("GetImportIDFn with an empty external name = %q, %v; want an empty id so upjet creates the resource", empty, err)
			}
			delete(params, "organization")
			if _, err := e.GetImportIDFn(ctx, c.externalName, params, nil); err == nil {
				t.Error("GetImportIDFn without the organization parameter must fail")
			}
		})
	}
}
