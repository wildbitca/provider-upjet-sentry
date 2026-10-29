package monitor

import (
	"github.com/crossplane/upjet/v2/pkg/config"

	"github.com/wildbitca/provider-upjet-sentry/config/common"
)

// group is the API short group of every resource in this package.
const group = "monitor"

// Configure adds the monitor group resource configurators. The group name and the
// Kind names are an API contract: renaming them breaks existing claims.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("sentry_cron_monitor", func(r *config.Resource) {
		r.ShortGroup = group
		r.Kind = "CronMonitor"
		r.References["organization"] = common.Organization()
		r.References["project"] = common.Project()
		r.References["owner.team_id"] = common.TeamInternalID()
	})

	p.AddResourceConfigurator("sentry_metric_monitor", func(r *config.Resource) {
		r.ShortGroup = group
		r.Kind = "MetricMonitor"
		r.References["organization"] = common.Organization()
		r.References["project"] = common.Project()
		r.References["owner.team_id"] = common.TeamInternalID()
	})

	p.AddResourceConfigurator("sentry_uptime_monitor", func(r *config.Resource) {
		r.ShortGroup = group
		r.Kind = "UptimeMonitor"
		r.References["organization"] = common.Organization()
		r.References["project"] = common.Project()
		r.References["owner.team_id"] = common.TeamInternalID()
	})
}
