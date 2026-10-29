package organization

import (
	"github.com/crossplane/upjet/v2/pkg/config"

	"github.com/wildbitca/provider-upjet-sentry/config/common"
)

// group is the API short group of every resource in this package.
const group = "organization"

// Configure adds the organization group resource configurators. The group name and the
// Kind names are an API contract: renaming them breaks existing claims.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("sentry_organization", func(r *config.Resource) {
		r.ShortGroup = group
		r.Kind = "Organization"
	})

	p.AddResourceConfigurator("sentry_organization_member", func(r *config.Resource) {
		r.ShortGroup = group
		r.Kind = "OrganizationMember"
		r.References["organization"] = common.Organization()
	})

	p.AddResourceConfigurator("sentry_organization_code_mapping", func(r *config.Resource) {
		r.ShortGroup = group
		r.Kind = "OrganizationCodeMapping"
		r.References["organization"] = common.Organization()
	})

	p.AddResourceConfigurator("sentry_organization_repository", func(r *config.Resource) {
		r.ShortGroup = group
		r.Kind = "OrganizationRepository"
		r.References["organization"] = common.Organization()
	})
}
