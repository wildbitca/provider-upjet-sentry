package integration

import (
	"github.com/crossplane/upjet/v2/pkg/config"

	"github.com/wildbitca/provider-upjet-sentry/config/common"
)

// group is the API short group of every resource in this package.
const group = "integration"

// Configure adds the integration group resource configurators. The group name and the
// Kind names are an API contract: renaming them breaks existing claims.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("sentry_integration_opsgenie", func(r *config.Resource) {
		r.ShortGroup = group
		r.Kind = "IntegrationOpsgenie"
		r.References["organization"] = common.Organization()
	})

	p.AddResourceConfigurator("sentry_integration_pagerduty", func(r *config.Resource) {
		r.ShortGroup = group
		r.Kind = "IntegrationPagerduty"
		r.References["organization"] = common.Organization()
	})
}
