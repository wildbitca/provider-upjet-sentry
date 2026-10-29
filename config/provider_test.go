package config

import (
	"testing"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// wantGroupKind is the public API contract: group and Kind names must never
// change after release because existing claims reference them.
var wantGroupKind = map[string][2]string{
	"sentry_organization":                  {"organization", "Organization"},
	"sentry_organization_member":           {"organization", "OrganizationMember"},
	"sentry_organization_code_mapping":     {"organization", "OrganizationCodeMapping"},
	"sentry_organization_repository":       {"organization", "OrganizationRepository"},
	"sentry_team":                          {"team", "Team"},
	"sentry_team_member":                   {"team", "TeamMember"},
	"sentry_project":                       {"project", "Project"},
	"sentry_key":                           {"project", "ProjectKey"},
	"sentry_plugin":                        {"project", "ProjectPlugin"},
	"sentry_project_ownership":             {"project", "ProjectOwnership"},
	"sentry_project_spike_protection":      {"project", "ProjectSpikeProtection"},
	"sentry_all_projects_spike_protection": {"project", "AllProjectsSpikeProtection"},
	"sentry_project_inbound_data_filter":   {"project", "ProjectInboundDataFilter"},
	"sentry_project_symbol_source":         {"project", "ProjectSymbolSource"},
	"sentry_issue_alert":                   {"alert", "IssueAlert"},
	"sentry_metric_alert":                  {"alert", "MetricAlert"},
	"sentry_alert":                         {"alert", "Alert"},
	"sentry_notification_action":           {"alert", "NotificationAction"},
	"sentry_cron_monitor":                  {"monitor", "CronMonitor"},
	"sentry_metric_monitor":                {"monitor", "MetricMonitor"},
	"sentry_uptime_monitor":                {"monitor", "UptimeMonitor"},
	"sentry_integration_opsgenie":          {"integration", "IntegrationOpsgenie"},
	"sentry_integration_pagerduty":         {"integration", "IntegrationPagerduty"},
	"sentry_dashboard":                     {"dashboard", "Dashboard"},
}

func providers() map[string]*ujconfig.Provider {
	return map[string]*ujconfig.Provider{
		"cluster":    GetProvider(),
		"namespaced": GetProviderNamespaced(),
	}
}

func TestGroupAndKindContract(t *testing.T) {
	if len(wantGroupKind) != len(wantResources) {
		t.Fatalf("contract table has %d entries, want %d", len(wantGroupKind), len(wantResources))
	}
	for scope, pc := range providers() {
		for name, want := range wantGroupKind {
			r, ok := pc.Resources[name]
			if !ok {
				t.Errorf("%s: resource %s was not generated", scope, name)
				continue
			}
			if r.ShortGroup != want[0] || r.Kind != want[1] {
				t.Errorf("%s: %s is %s/%s, want %s/%s", scope, name, r.ShortGroup, r.Kind, want[0], want[1])
			}
		}
		if len(pc.Resources) != len(wantGroupKind) {
			t.Errorf("%s: %d resources generated, want %d", scope, len(pc.Resources), len(wantGroupKind))
		}
	}
}

func TestDeprecatedAttributesAreOmitted(t *testing.T) {
	cases := map[string][]string{
		"sentry_team":        {"team_id"},
		"sentry_key":         {"dsn_csp", "dsn_public", "dsn_secret"},
		"sentry_issue_alert": {"actions", "conditions", "filters"},
	}
	for scope, pc := range providers() {
		for name, attrs := range cases {
			for _, a := range attrs {
				if _, ok := pc.Resources[name].TerraformResource.Schema[a]; ok {
					t.Errorf("%s: %s.%s must be omitted", scope, name, a)
				}
			}
		}
		af := pc.Resources["sentry_alert"].TerraformResource.Schema["action_filters"].Elem.(*schema.Resource)
		actions := af.Schema["actions"].Elem.(*schema.Resource)
		if _, ok := actions.Schema["plugin"]; ok {
			t.Errorf("%s: sentry_alert.action_filters.actions.plugin must be omitted", scope)
		}
	}
}

// wantReferences is the cross-resource reference contract: Terraform field path
// -> Terraform name of the referenced resource, per resource. Only these fields
// are references; everything else (integration ids, monitor ids, user ids, the
// Opsgenie team label, the code mapping project id) stays a plain value.
var wantReferences = map[string]map[string]string{
	"sentry_organization":                  nil,
	"sentry_organization_member":           {"organization": "sentry_organization"},
	"sentry_organization_code_mapping":     {"organization": "sentry_organization"},
	"sentry_organization_repository":       {"organization": "sentry_organization"},
	"sentry_team":                          {"organization": "sentry_organization"},
	"sentry_team_member":                   {"organization": "sentry_organization", "team": "sentry_team", "member_id": "sentry_organization_member"},
	"sentry_project":                       {"organization": "sentry_organization", "teams": "sentry_team"},
	"sentry_key":                           {"organization": "sentry_organization", "project": "sentry_project"},
	"sentry_plugin":                        {"organization": "sentry_organization", "project": "sentry_project"},
	"sentry_project_ownership":             {"organization": "sentry_organization", "project": "sentry_project"},
	"sentry_project_spike_protection":      {"organization": "sentry_organization", "project": "sentry_project"},
	"sentry_all_projects_spike_protection": {"organization": "sentry_organization", "projects": "sentry_project"},
	"sentry_project_inbound_data_filter":   {"organization": "sentry_organization", "project": "sentry_project"},
	"sentry_project_symbol_source":         {"organization": "sentry_organization", "project": "sentry_project"},
	"sentry_issue_alert":                   {"organization": "sentry_organization", "project": "sentry_project"},
	"sentry_metric_alert":                  {"organization": "sentry_organization", "project": "sentry_project"},
	"sentry_alert":                         {"organization": "sentry_organization"},
	"sentry_notification_action":           {"organization": "sentry_organization", "projects": "sentry_project"},
	"sentry_cron_monitor":                  {"organization": "sentry_organization", "project": "sentry_project", "owner.team_id": "sentry_team"},
	"sentry_metric_monitor":                {"organization": "sentry_organization", "project": "sentry_project", "owner.team_id": "sentry_team"},
	"sentry_uptime_monitor":                {"organization": "sentry_organization", "project": "sentry_project", "owner.team_id": "sentry_team"},
	"sentry_integration_opsgenie":          {"organization": "sentry_organization"},
	"sentry_integration_pagerduty":         {"organization": "sentry_organization"},
	"sentry_dashboard":                     {"organization": "sentry_organization"},
}

func TestReferences(t *testing.T) {
	for scope, pc := range providers() {
		for name, want := range wantReferences {
			got := pc.Resources[name].References
			if len(got) != len(want) {
				t.Errorf("%s: %s has references %v, want %v", scope, name, keys(got), want)
			}
			for field, target := range want {
				ref, ok := got[field]
				if !ok {
					t.Errorf("%s: %s.%s has no reference", scope, name, field)
					continue
				}
				if ref.TerraformName != target {
					t.Errorf("%s: %s.%s references %s, want %s", scope, name, field, ref.TerraformName, target)
				}
			}
		}
	}
}

func keys(m ujconfig.References) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// TestReferenceGroupsAreAcyclic guards the Go import graph of the generated
// API packages: a reference creates an import from the referencing group to the
// referenced one, so the group graph must stay acyclic
// (organization <- team <- project <- alert/monitor, dashboard/integration leaf).
func TestReferenceGroupsAreAcyclic(t *testing.T) {
	pc := GetProvider()
	deps := map[string]map[string]bool{}
	for name, refs := range wantReferences {
		from := wantGroupKind[name][0]
		for field := range refs {
			ref := pc.Resources[name].References[field]
			to := wantGroupKind[ref.TerraformName][0]
			if from == to {
				continue
			}
			if deps[from] == nil {
				deps[from] = map[string]bool{}
			}
			deps[from][to] = true
		}
	}
	const (
		visiting = 1
		done     = 2
	)
	state := map[string]int{}
	var visit func(g string, path []string)
	visit = func(g string, path []string) {
		switch state[g] {
		case visiting:
			t.Fatalf("group reference cycle: %v -> %s", path, g)
		case done:
			return
		}
		state[g] = visiting
		for next := range deps[g] {
			visit(next, append(path, g))
		}
		state[g] = done
	}
	for g := range deps {
		visit(g, nil)
	}
}
