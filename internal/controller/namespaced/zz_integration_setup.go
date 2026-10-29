// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	integrationopsgenie "github.com/wildbitca/provider-upjet-sentry/internal/controller/namespaced/integration/integrationopsgenie"
	integrationpagerduty "github.com/wildbitca/provider-upjet-sentry/internal/controller/namespaced/integration/integrationpagerduty"
)

// Setup_integration creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup_integration(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		integrationopsgenie.Setup,
		integrationpagerduty.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupGated_integration creates all controllers with the supplied logger and adds them to
// the supplied manager gated.
func SetupGated_integration(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		integrationopsgenie.SetupGated,
		integrationpagerduty.SetupGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupWebhookWithManager_integration registers conversion webhooks for all resource kinds in the group.
func SetupWebhookWithManager_integration(mgr ctrl.Manager) error {
	for _, setup := range []func(ctrl.Manager) error{
		integrationopsgenie.SetupWebhookWithManager,
		integrationpagerduty.SetupWebhookWithManager,
	} {
		if err := setup(mgr); err != nil {
			return err
		}
	}
	return nil
}
