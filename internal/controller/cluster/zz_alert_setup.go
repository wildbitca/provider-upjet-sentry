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
)

// Setup_alert creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup_alert(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		alert.Setup,
		issuealert.Setup,
		metricalert.Setup,
		notificationaction.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupGated_alert creates all controllers with the supplied logger and adds them to
// the supplied manager gated.
func SetupGated_alert(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		alert.SetupGated,
		issuealert.SetupGated,
		metricalert.SetupGated,
		notificationaction.SetupGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
