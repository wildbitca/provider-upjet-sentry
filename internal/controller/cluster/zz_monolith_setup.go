// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	alert "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/alert/alert"
	issuealert "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/alert/issuealert"
	metricalert "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/alert/metricalert"
	notificationaction "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/alert/notificationaction"
	dashboard "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/dashboard/dashboard"
	integrationopsgenie "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/integration/integrationopsgenie"
	integrationpagerduty "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/integration/integrationpagerduty"
	cronmonitor "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/monitor/cronmonitor"
	metricmonitor "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/monitor/metricmonitor"
	uptimemonitor "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/monitor/uptimemonitor"
	organization "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/organization/organization"
	organizationcodemapping "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/organization/organizationcodemapping"
	organizationmember "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/organization/organizationmember"
	organizationrepository "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/organization/organizationrepository"
	allprojectsspikeprotection "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/project/allprojectsspikeprotection"
	project "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/project/project"
	projectinbounddatafilter "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/project/projectinbounddatafilter"
	projectkey "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/project/projectkey"
	projectownership "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/project/projectownership"
	projectplugin "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/project/projectplugin"
	projectspikeprotection "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/project/projectspikeprotection"
	projectsymbolsource "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/project/projectsymbolsource"
	providerconfig "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/providerconfig"
	team "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/team/team"
	teammember "github.com/wildbitca/provider-upjet-sentry/internal/controller/cluster/team/teammember"
)

// Setup_monolith creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup_monolith(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		alert.Setup,
		issuealert.Setup,
		metricalert.Setup,
		notificationaction.Setup,
		dashboard.Setup,
		integrationopsgenie.Setup,
		integrationpagerduty.Setup,
		cronmonitor.Setup,
		metricmonitor.Setup,
		uptimemonitor.Setup,
		organization.Setup,
		organizationcodemapping.Setup,
		organizationmember.Setup,
		organizationrepository.Setup,
		allprojectsspikeprotection.Setup,
		project.Setup,
		projectinbounddatafilter.Setup,
		projectkey.Setup,
		projectownership.Setup,
		projectplugin.Setup,
		projectspikeprotection.Setup,
		projectsymbolsource.Setup,
		providerconfig.Setup,
		team.Setup,
		teammember.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupGated_monolith creates all controllers with the supplied logger and adds them to
// the supplied manager gated.
func SetupGated_monolith(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		alert.SetupGated,
		issuealert.SetupGated,
		metricalert.SetupGated,
		notificationaction.SetupGated,
		dashboard.SetupGated,
		integrationopsgenie.SetupGated,
		integrationpagerduty.SetupGated,
		cronmonitor.SetupGated,
		metricmonitor.SetupGated,
		uptimemonitor.SetupGated,
		organization.SetupGated,
		organizationcodemapping.SetupGated,
		organizationmember.SetupGated,
		organizationrepository.SetupGated,
		allprojectsspikeprotection.SetupGated,
		project.SetupGated,
		projectinbounddatafilter.SetupGated,
		projectkey.SetupGated,
		projectownership.SetupGated,
		projectplugin.SetupGated,
		projectspikeprotection.SetupGated,
		projectsymbolsource.SetupGated,
		providerconfig.SetupGated,
		team.SetupGated,
		teammember.SetupGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupWebhookWithManager_monolith registers conversion webhooks for all resource kinds in the group.
func SetupWebhookWithManager_monolith(mgr ctrl.Manager) error {
	for _, setup := range []func(ctrl.Manager) error{
		alert.SetupWebhookWithManager,
		issuealert.SetupWebhookWithManager,
		metricalert.SetupWebhookWithManager,
		notificationaction.SetupWebhookWithManager,
		dashboard.SetupWebhookWithManager,
		integrationopsgenie.SetupWebhookWithManager,
		integrationpagerduty.SetupWebhookWithManager,
		cronmonitor.SetupWebhookWithManager,
		metricmonitor.SetupWebhookWithManager,
		uptimemonitor.SetupWebhookWithManager,
		organization.SetupWebhookWithManager,
		organizationcodemapping.SetupWebhookWithManager,
		organizationmember.SetupWebhookWithManager,
		organizationrepository.SetupWebhookWithManager,
		allprojectsspikeprotection.SetupWebhookWithManager,
		project.SetupWebhookWithManager,
		projectinbounddatafilter.SetupWebhookWithManager,
		projectkey.SetupWebhookWithManager,
		projectownership.SetupWebhookWithManager,
		projectplugin.SetupWebhookWithManager,
		projectspikeprotection.SetupWebhookWithManager,
		projectsymbolsource.SetupWebhookWithManager,
		providerconfig.SetupWebhookWithManager,
		team.SetupWebhookWithManager,
		teammember.SetupWebhookWithManager,
	} {
		if err := setup(mgr); err != nil {
			return err
		}
	}
	return nil
}
