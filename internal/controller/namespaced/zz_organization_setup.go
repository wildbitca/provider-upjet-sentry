// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	organization "github.com/wildbitca/provider-upjet-sentry/internal/controller/namespaced/organization/organization"
	organizationcodemapping "github.com/wildbitca/provider-upjet-sentry/internal/controller/namespaced/organization/organizationcodemapping"
	organizationmember "github.com/wildbitca/provider-upjet-sentry/internal/controller/namespaced/organization/organizationmember"
	organizationrepository "github.com/wildbitca/provider-upjet-sentry/internal/controller/namespaced/organization/organizationrepository"
)

// Setup_organization creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup_organization(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		organization.Setup,
		organizationcodemapping.Setup,
		organizationmember.Setup,
		organizationrepository.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupGated_organization creates all controllers with the supplied logger and adds them to
// the supplied manager gated.
func SetupGated_organization(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		organization.SetupGated,
		organizationcodemapping.SetupGated,
		organizationmember.SetupGated,
		organizationrepository.SetupGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
