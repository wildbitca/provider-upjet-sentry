package config

import (
	"context"
	"strings"
	"testing"
)

// wantResources is the contract: every managed resource of the pinned
// jianyuan/sentry release must be configured, otherwise it silently does not
// compile (ExternalNameConfigs doubles as the include list).
var wantResources = []string{
	"sentry_alert",
	"sentry_all_projects_spike_protection",
	"sentry_cron_monitor",
	"sentry_dashboard",
	"sentry_integration_opsgenie",
	"sentry_integration_pagerduty",
	"sentry_issue_alert",
	"sentry_key",
	"sentry_metric_alert",
	"sentry_metric_monitor",
	"sentry_notification_action",
	"sentry_organization",
	"sentry_organization_code_mapping",
	"sentry_organization_member",
	"sentry_organization_repository",
	"sentry_plugin",
	"sentry_project",
	"sentry_project_inbound_data_filter",
	"sentry_project_ownership",
	"sentry_project_spike_protection",
	"sentry_project_symbol_source",
	"sentry_team",
	"sentry_team_member",
	"sentry_uptime_monitor",
}

func TestIncludeListHasAllResources(t *testing.T) {
	got := map[string]bool{}
	for _, n := range ExternalNameConfigured() {
		got[n] = true
	}
	if len(got) != len(wantResources) {
		t.Fatalf("include list has %d resources, want %d", len(got), len(wantResources))
	}
	for _, n := range wantResources {
		if !got[n] {
			t.Errorf("resource %s is missing from the include list", n)
		}
	}
}

func TestGetIDFn(t *testing.T) {
	ctx := context.Background()
	cases := map[string]struct {
		resource string
		en       string
		params   map[string]any
		want     string
		wantErr  string
	}{
		"OrganizationMemberComposite": {
			resource: "sentry_organization_member", en: "42",
			params: map[string]any{"organization": "acme"}, want: "acme/42",
		},
		"MetricAlertComposite": {
			resource: "sentry_metric_alert", en: "7",
			params: map[string]any{"organization": "acme", "project": "web"}, want: "acme/web/7",
		},
		"DashboardComposite": {
			resource: "sentry_dashboard", en: "9",
			params: map[string]any{"organization": "acme"}, want: "acme/9",
		},
		"KeyUsesBareIDInState": {
			resource: "sentry_key", en: "abc",
			params: map[string]any{"organization": "acme", "project": "web"}, want: "abc",
		},
		"IssueAlertUsesBareIDInState": {
			resource: "sentry_issue_alert", en: "55",
			params: map[string]any{"organization": "acme", "project": "web"}, want: "55",
		},
		"TeamIsSlug": {
			resource: "sentry_team", en: "backend",
			params: map[string]any{"organization": "acme"}, want: "backend",
		},
		"TeamMemberFromParams": {
			resource: "sentry_team_member", en: "ignored",
			params: map[string]any{"organization": "acme", "team": "backend", "member_id": "11"}, want: "acme/backend/11",
		},
		"TeamMemberMissingParam": {
			resource: "sentry_team_member", en: "11",
			params: map[string]any{"organization": "acme"}, wantErr: "team",
		},
		"InboundDataFilterFromParams": {
			resource: "sentry_project_inbound_data_filter", en: "x",
			params: map[string]any{"organization": "acme", "project": "web", "filter_id": "browser-extensions"}, want: "acme/web/browser-extensions",
		},
		"ProjectSpikeProtectionFromParams": {
			resource: "sentry_project_spike_protection", en: "x",
			params: map[string]any{"organization": "acme", "project": "web"}, want: "acme/web",
		},
		"ProjectOwnershipImportString": {
			resource: "sentry_project_ownership", en: "",
			params: map[string]any{"organization": "acme", "project": "web"}, want: "acme/web",
		},
		"AllProjectsSpikeProtectionImportString": {
			resource: "sentry_all_projects_spike_protection", en: "",
			params: map[string]any{"organization": "acme"}, want: "acme",
		},
		"OwnershipMissingProject": {
			resource: "sentry_project_ownership", en: "",
			params: map[string]any{"organization": "acme"}, wantErr: "project",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, ok := ExternalNameConfigs[tc.resource]
			if !ok {
				t.Fatalf("no external name config for %s", tc.resource)
			}
			got, err := e.GetIDFn(ctx, tc.en, tc.params, nil)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("want error containing %q, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("GetIDFn = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestGetImportIDFn(t *testing.T) {
	ctx := context.Background()
	cases := map[string]struct {
		resource string
		en       string
		params   map[string]any
		want     string
	}{
		"TeamIsOrgSlashSlug":     {"sentry_team", "backend", map[string]any{"organization": "acme"}, "acme/backend"},
		"ProjectIsOrgSlashSlug":  {"sentry_project", "web", map[string]any{"organization": "acme"}, "acme/web"},
		"KeyIsOrgProjectID":      {"sentry_key", "abc", map[string]any{"organization": "acme", "project": "web"}, "acme/web/abc"},
		"OpsgenieHasIntegration": {"sentry_integration_opsgenie", "svc", map[string]any{"organization": "acme", "integration_id": "5"}, "acme/5/svc"},
		"RepositoryHasFourParts": {"sentry_organization_repository", "88", map[string]any{"organization": "acme", "integration_type": "github", "integration_id": "12"}, "acme/github/12/88"},
		"PluginIsOrgProjectName": {"sentry_plugin", "slack", map[string]any{"organization": "acme", "project": "web", "plugin": "slack"}, "acme/web/slack"},
		"NoExternalNameYet":      {"sentry_alert", "", map[string]any{"organization": "acme"}, ""},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := ExternalNameConfigs[tc.resource].GetImportIDFn(ctx, tc.en, tc.params, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("GetImportIDFn = %q, want %q", got, tc.want)
			}
		})
	}
	if _, err := ExternalNameConfigs["sentry_issue_alert"].GetImportIDFn(ctx, "55", map[string]any{"organization": "acme"}, nil); err == nil || !strings.Contains(err.Error(), "project") {
		t.Fatalf("want an error naming the missing project, got %v", err)
	}
}

func TestGetExternalNameFn(t *testing.T) {
	cases := map[string]struct {
		resource string
		tfstate  map[string]any
		want     string
		wantErr  string
	}{
		"OrganizationMemberFromCompositeID": {
			resource: "sentry_organization_member",
			tfstate:  map[string]any{"id": "acme/42"}, want: "42",
		},
		"MetricAlertFromCompositeID": {
			resource: "sentry_metric_alert",
			tfstate:  map[string]any{"id": "acme/web/7"}, want: "7",
		},
		"DashboardFromCompositeID": {
			resource: "sentry_dashboard",
			tfstate:  map[string]any{"id": "acme/9"}, want: "9",
		},
		"KeyFromBareID": {
			resource: "sentry_key",
			tfstate:  map[string]any{"id": "abc"}, want: "abc",
		},
		"KeyMissingID": {
			resource: "sentry_key",
			tfstate:  map[string]any{}, wantErr: "id",
		},
		"TeamMemberIsMemberID": {
			resource: "sentry_team_member",
			tfstate:  map[string]any{"id": "acme/backend/11", "member_id": "11"}, want: "11",
		},
		"InboundDataFilterIsFilterID": {
			resource: "sentry_project_inbound_data_filter",
			tfstate:  map[string]any{"filter_id": "legacy-browsers"}, want: "legacy-browsers",
		},
		"ProjectSpikeProtectionIsProjectSlug": {
			resource: "sentry_project_spike_protection",
			tfstate:  map[string]any{"organization": "acme", "project": "web"}, want: "web",
		},
		"OwnershipWithoutIDUsesParams": {
			resource: "sentry_project_ownership",
			tfstate:  map[string]any{"organization": "acme", "project": "web"}, want: "acme/web",
		},
		"AllProjectsWithoutIDUsesOrganization": {
			resource: "sentry_all_projects_spike_protection",
			tfstate:  map[string]any{"organization": "acme"}, want: "acme",
		},
		"OwnershipMissingAttributes": {
			resource: "sentry_project_ownership",
			tfstate:  map[string]any{}, wantErr: "organization",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, ok := ExternalNameConfigs[tc.resource]
			if !ok {
				t.Fatalf("no external name config for %s", tc.resource)
			}
			got, err := e.GetExternalNameFn(tc.tfstate)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("want error containing %q, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("GetExternalNameFn = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSlugResourcesSeedSlugFromExternalName(t *testing.T) {
	for _, r := range []string{"sentry_organization", "sentry_team", "sentry_project"} {
		e := ExternalNameConfigs[r]
		base := map[string]any{}
		e.SetIdentifierArgumentFn(base, "my-slug")
		if base["slug"] != "my-slug" {
			t.Errorf("%s: slug = %v, want my-slug", r, base["slug"])
		}
		if len(e.OmittedFields) != 1 || e.OmittedFields[0] != "slug" {
			t.Errorf("%s: OmittedFields = %v, want [slug]", r, e.OmittedFields)
		}
	}
}
