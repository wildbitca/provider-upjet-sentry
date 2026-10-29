package project

import (
	"github.com/crossplane/upjet/v2/pkg/config"

	"github.com/wildbitca/provider-upjet-sentry/config/common"
)

// group is the API short group of every resource in this package.
const group = "project"

// Configure adds the project group resource configurators. The group name and the
// Kind names are an API contract: renaming them breaks existing claims.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("sentry_project", func(r *config.Resource) {
		r.ShortGroup = group
		r.Kind = "Project"
		r.References["organization"] = common.Organization()
		r.References["teams"] = common.Team()
	})

	p.AddResourceConfigurator("sentry_key", func(r *config.Resource) {
		r.ShortGroup = group
		r.Kind = "ProjectKey"
		r.References["organization"] = common.Organization()
		r.References["project"] = common.Project()
		common.Omit(r, "dsn_csp", "dsn_public", "dsn_secret") // deprecated: use the dsn map
	})

	p.AddResourceConfigurator("sentry_plugin", func(r *config.Resource) {
		r.ShortGroup = group
		r.Kind = "ProjectPlugin"
		r.References["organization"] = common.Organization()
		r.References["project"] = common.Project()
	})

	p.AddResourceConfigurator("sentry_project_ownership", func(r *config.Resource) {
		r.ShortGroup = group
		r.Kind = "ProjectOwnership"
		r.References["organization"] = common.Organization()
		r.References["project"] = common.Project()
	})

	p.AddResourceConfigurator("sentry_project_spike_protection", func(r *config.Resource) {
		r.ShortGroup = group
		r.Kind = "ProjectSpikeProtection"
		r.References["organization"] = common.Organization()
		r.References["project"] = common.Project()
	})

	p.AddResourceConfigurator("sentry_all_projects_spike_protection", func(r *config.Resource) {
		r.ShortGroup = group
		r.Kind = "AllProjectsSpikeProtection"
		r.References["organization"] = common.Organization()
		r.References["projects"] = common.Project()
	})

	p.AddResourceConfigurator("sentry_project_inbound_data_filter", func(r *config.Resource) {
		r.ShortGroup = group
		r.Kind = "ProjectInboundDataFilter"
		r.References["organization"] = common.Organization()
		r.References["project"] = common.Project()
	})

	p.AddResourceConfigurator("sentry_project_symbol_source", func(r *config.Resource) {
		r.ShortGroup = group
		r.Kind = "ProjectSymbolSource"
		r.References["organization"] = common.Organization()
		r.References["project"] = common.Project()
	})
}
