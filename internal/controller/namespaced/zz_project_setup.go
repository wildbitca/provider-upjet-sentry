// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	allprojectsspikeprotection "github.com/wildbitca/provider-upjet-sentry/internal/controller/namespaced/project/allprojectsspikeprotection"
	project "github.com/wildbitca/provider-upjet-sentry/internal/controller/namespaced/project/project"
	projectinbounddatafilter "github.com/wildbitca/provider-upjet-sentry/internal/controller/namespaced/project/projectinbounddatafilter"
	projectkey "github.com/wildbitca/provider-upjet-sentry/internal/controller/namespaced/project/projectkey"
	projectownership "github.com/wildbitca/provider-upjet-sentry/internal/controller/namespaced/project/projectownership"
	projectplugin "github.com/wildbitca/provider-upjet-sentry/internal/controller/namespaced/project/projectplugin"
	projectspikeprotection "github.com/wildbitca/provider-upjet-sentry/internal/controller/namespaced/project/projectspikeprotection"
	projectsymbolsource "github.com/wildbitca/provider-upjet-sentry/internal/controller/namespaced/project/projectsymbolsource"
)

// Setup_project creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup_project(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		allprojectsspikeprotection.Setup,
		project.Setup,
		projectinbounddatafilter.Setup,
		projectkey.Setup,
		projectownership.Setup,
		projectplugin.Setup,
		projectspikeprotection.Setup,
		projectsymbolsource.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupGated_project creates all controllers with the supplied logger and adds them to
// the supplied manager gated.
func SetupGated_project(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		allprojectsspikeprotection.SetupGated,
		project.SetupGated,
		projectinbounddatafilter.SetupGated,
		projectkey.SetupGated,
		projectownership.SetupGated,
		projectplugin.SetupGated,
		projectspikeprotection.SetupGated,
		projectsymbolsource.SetupGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupWebhookWithManager_project registers conversion webhooks for all resource kinds in the group.
func SetupWebhookWithManager_project(mgr ctrl.Manager) error {
	for _, setup := range []func(ctrl.Manager) error{
		allprojectsspikeprotection.SetupWebhookWithManager,
		project.SetupWebhookWithManager,
		projectinbounddatafilter.SetupWebhookWithManager,
		projectkey.SetupWebhookWithManager,
		projectownership.SetupWebhookWithManager,
		projectplugin.SetupWebhookWithManager,
		projectspikeprotection.SetupWebhookWithManager,
		projectsymbolsource.SetupWebhookWithManager,
	} {
		if err := setup(mgr); err != nil {
			return err
		}
	}
	return nil
}
