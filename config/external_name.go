package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/crossplane/upjet/v2/pkg/config"
)

// Spike S1 outcome (docs/architecture/spikes.md): upjet writes the output of
// GetIDFn into the synthetic Terraform state `id` before every refresh, so
// GetIDFn must reproduce exactly the `id` the provider itself stores. Resources
// that store a bare id in state therefore keep the bare id (the organization
// and project attributes that Read needs are copied from the parameters), and
// only resources whose state id is a composite use a composite.

// slugAsIdentifier is used by resources addressed by slug (organization, team,
// project). The slug is the Terraform id, the import key and the external
// name, and it is seeded from the external name so that adoption by name works.
var slugAsIdentifier = config.ExternalName{
	SetIdentifierArgumentFn: func(base map[string]any, externalName string) {
		base["slug"] = externalName
	},
	GetExternalNameFn:      config.IDAsExternalName,
	GetIDFn:                config.ExternalNameAsID,
	OmittedFields:          []string{"slug"},
	DisableNameInitializer: false,
}

// compositeFromProvider covers resources whose state id is a composite
// "<parameters...>/<provider id>". template is a Go text/template over
// .parameters and .external_name.
func compositeFromProvider(template string) config.ExternalName {
	e := config.TemplatedStringAsIdentifier("", template)
	e.DisableNameInitializer = true
	return e
}

// paramDrivenID covers resources whose identity is made only of their own
// parameters (there is nothing the provider assigns). idKeys are joined with
// "/" to form the Terraform id or import string; the external name is the
// attribute at nameKeys joined with "/".
func paramDrivenID(idKeys, nameKeys []string) config.ExternalName {
	return config.ExternalName{
		SetIdentifierArgumentFn: config.NopSetIdentifierArgument,
		GetExternalNameFn: func(tfstate map[string]any) (string, error) {
			return joinStrings(tfstate, nameKeys, "tfstate")
		},
		GetIDFn: func(_ context.Context, _ string, parameters, _ map[string]any) (string, error) {
			return joinStrings(parameters, idKeys, "parameters")
		},
		DisableNameInitializer: true,
	}
}

func joinStrings(m map[string]any, keys []string, what string) (string, error) {
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		v, ok := m[k].(string)
		if !ok || v == "" {
			return "", fmt.Errorf("cannot determine external name: %q is required in %s", k, what)
		}
		parts = append(parts, v)
	}
	return strings.Join(parts, "/"), nil
}

// withImportID sets GetImportIDFn on e. The Terraform import string is
// "<parameters...>/<external name>", with the parameters read in the order of
// keys, while the state id (GetIDFn) is left untouched.
//
// Upjet 2.4.0+ passes GetImportIDFn, not GetIDFn, to `terraform import` (the
// Observe-only and adoption path). The jianyuan/sentry importers of the
// resources below require a composite string even though the id they store in
// state is bare, so importing with the bare id fails ("expected format
// org/id"). An empty external name yields an empty import id, which upjet
// reads as "the resource does not exist yet" and routes to Create.
func withImportID(e config.ExternalName, keys ...string) config.ExternalName {
	e.GetImportIDFn = func(_ context.Context, externalName string, parameters, _ map[string]any) (string, error) {
		if externalName == "" {
			return "", nil
		}
		parts := make([]string, 0, len(keys)+1)
		for _, k := range keys {
			v, ok := parameters[k].(string)
			if !ok || v == "" {
				return "", fmt.Errorf("cannot build the import id: %q is required in parameters", k)
			}
			parts = append(parts, v)
		}
		return strings.Join(append(parts, externalName), "/"), nil
	}
	return e
}

// ExternalNameConfigs contains all external name configurations for this
// provider and doubles as the include list. jianyuan/sentry v0.15.8 has 24
// managed resources; see docs/architecture/external-names.md for the classes.
var ExternalNameConfigs = map[string]config.ExternalName{
	// Slug addressed.
	"sentry_organization": slugAsIdentifier,
	"sentry_team":         withImportID(slugAsIdentifier, "organization"),
	"sentry_project":      withImportID(slugAsIdentifier, "organization"),

	// Provider assigned id, stored bare in state.
	"sentry_organization_code_mapping": withImportID(config.IdentifierFromProvider, "organization"),
	"sentry_organization_repository":   withImportID(config.IdentifierFromProvider, "organization", "integration_type", "integration_id"),
	"sentry_key":                       withImportID(config.IdentifierFromProvider, "organization", "project"),
	"sentry_project_symbol_source":     withImportID(config.IdentifierFromProvider, "organization", "project"),
	"sentry_issue_alert":               withImportID(config.IdentifierFromProvider, "organization", "project"),
	"sentry_alert":                     withImportID(config.IdentifierFromProvider, "organization"),
	"sentry_notification_action":       withImportID(config.IdentifierFromProvider, "organization"),
	"sentry_cron_monitor":              withImportID(config.IdentifierFromProvider, "organization"),
	"sentry_metric_monitor":            withImportID(config.IdentifierFromProvider, "organization"),
	"sentry_uptime_monitor":            withImportID(config.IdentifierFromProvider, "organization"),
	"sentry_integration_opsgenie":      withImportID(config.IdentifierFromProvider, "organization", "integration_id"),
	"sentry_integration_pagerduty":     withImportID(config.IdentifierFromProvider, "organization", "integration_id"),

	// Provider assigned id, stored as a composite in state.
	"sentry_organization_member": compositeFromProvider("{{ .parameters.organization }}/{{ .external_name }}"),
	"sentry_metric_alert":        compositeFromProvider("{{ .parameters.organization }}/{{ .parameters.project }}/{{ .external_name }}"),
	"sentry_dashboard":           compositeFromProvider("{{ .parameters.organization }}/{{ .external_name }}"),

	// Identity made only of parameters. A plugin's id is its name, which is the
	// required `plugin` parameter.
	"sentry_plugin": withImportID(paramDrivenID([]string{"plugin"}, []string{"plugin"}), "organization", "project"),
	"sentry_team_member": paramDrivenID(
		[]string{"organization", "team", "member_id"}, []string{"member_id"}),
	"sentry_project_inbound_data_filter": paramDrivenID(
		[]string{"organization", "project", "filter_id"}, []string{"filter_id"}),
	"sentry_project_spike_protection": paramDrivenID(
		[]string{"organization", "project"}, []string{"project"}),

	// No `id` attribute at all: identity is the import string.
	"sentry_project_ownership": paramDrivenID(
		[]string{"organization", "project"}, []string{"organization", "project"}),
	// Note: sentry_all_projects_spike_protection has no Terraform resource importer
	// ("Resource Import Not Implemented"), so Observe-only import cannot succeed.
	// Retained here for parameter-driven external name configuration when managed with Create/Update.
	"sentry_all_projects_spike_protection": paramDrivenID(
		[]string{"organization"}, []string{"organization"}),
}

// ExternalNameConfigurations returns the list of all ExternalName configured resources.
func ExternalNameConfigurations() config.ResourceOption {
	return func(r *config.Resource) {
		if e, ok := ExternalNameConfigs[r.Name]; ok {
			r.ExternalName = e
		}
	}
}

// ExternalNameConfigured returns a list of all resources with external name configured.
func ExternalNameConfigured() []string {
	l := make([]string, 0, len(ExternalNameConfigs))
	for name := range ExternalNameConfigs {
		l = append(l, name)
	}
	return l
}
