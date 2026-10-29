// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	dashboard "github.com/wildbitca/provider-upjet-sentry/internal/controller/namespaced/dashboard/dashboard"
)

// Setup_dashboard creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup_dashboard(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		dashboard.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupGated_dashboard creates all controllers with the supplied logger and adds them to
// the supplied manager gated.
func SetupGated_dashboard(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		dashboard.SetupGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupWebhookWithManager_dashboard registers conversion webhooks for all resource kinds in the group.
func SetupWebhookWithManager_dashboard(mgr ctrl.Manager) error {
	for _, setup := range []func(ctrl.Manager) error{
		dashboard.SetupWebhookWithManager,
	} {
		if err := setup(mgr); err != nil {
			return err
		}
	}
	return nil
}
