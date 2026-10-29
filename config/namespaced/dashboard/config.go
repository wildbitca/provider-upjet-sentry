package dashboard

import (
	"github.com/crossplane/upjet/v2/pkg/config"

	"github.com/wildbitca/provider-upjet-sentry/config/common"
)

// group is the API short group of every resource in this package.
const group = "dashboard"

// Configure adds the dashboard group resource configurators. The group name and the
// Kind names are an API contract: renaming them breaks existing claims.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("sentry_dashboard", func(r *config.Resource) {
		r.ShortGroup = group
		r.Kind = "Dashboard"
		r.References["organization"] = common.Organization()
	})
}
