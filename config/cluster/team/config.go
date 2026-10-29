package team

import (
	"github.com/crossplane/upjet/v2/pkg/config"

	"github.com/wildbitca/provider-upjet-sentry/config/common"
)

// group is the API short group of every resource in this package.
const group = "team"

// Configure adds the team group resource configurators. The group name and the
// Kind names are an API contract: renaming them breaks existing claims.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("sentry_team", func(r *config.Resource) {
		r.ShortGroup = group
		r.Kind = "Team"
		r.References["organization"] = common.Organization()
		common.Omit(r, "team_id") // deprecated: use internal_id
	})

	p.AddResourceConfigurator("sentry_team_member", func(r *config.Resource) {
		r.ShortGroup = group
		r.Kind = "TeamMember"
		r.References["organization"] = common.Organization()
		r.References["team"] = common.Team()
		r.References["member_id"] = common.OrganizationMember()
	})
}
