package alert

import (
	"github.com/crossplane/upjet/v2/pkg/config"

	"github.com/wildbitca/provider-upjet-sentry/config/common"
)

// group is the API short group of every resource in this package.
const group = "alert"

// Configure adds the alert group resource configurators. The group name and the
// Kind names are an API contract: renaming them breaks existing claims.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("sentry_issue_alert", func(r *config.Resource) {
		r.ShortGroup = group
		r.Kind = "IssueAlert"
		r.References["organization"] = common.Organization()
		r.References["project"] = common.Project()
		common.Omit(r, "actions", "conditions", "filters") // deprecated JSON strings: use the *_v2 attributes
	})

	p.AddResourceConfigurator("sentry_metric_alert", func(r *config.Resource) {
		r.ShortGroup = group
		r.Kind = "MetricAlert"
		r.References["organization"] = common.Organization()
		r.References["project"] = common.Project()
	})

	p.AddResourceConfigurator("sentry_alert", func(r *config.Resource) {
		r.ShortGroup = group
		r.Kind = "Alert"
		r.References["organization"] = common.Organization()
		common.Omit(r, "action_filters.actions.plugin") // deprecated: cannot be created
	})

	p.AddResourceConfigurator("sentry_notification_action", func(r *config.Resource) {
		r.ShortGroup = group
		r.Kind = "NotificationAction"
		r.References["organization"] = common.Organization()
		r.References["projects"] = common.Project()
	})
}
