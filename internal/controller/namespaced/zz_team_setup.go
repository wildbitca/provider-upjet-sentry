// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	team "github.com/wildbitca/provider-upjet-sentry/internal/controller/namespaced/team/team"
	teammember "github.com/wildbitca/provider-upjet-sentry/internal/controller/namespaced/team/teammember"
)

// Setup_team creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup_team(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		team.Setup,
		teammember.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupGated_team creates all controllers with the supplied logger and adds them to
// the supplied manager gated.
func SetupGated_team(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		team.SetupGated,
		teammember.SetupGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
