//go:build generate
// +build generate

// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// NOTE: See the below link for details on what is happening here.
// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

// Remove any previously generated CRDs so that removed API groups do not leave
// stale manifests behind.
//go:generate rm -rf ../package/crds

// Run the Upjet code-generation pipeline. This generates the API types,
// terraformed implementations, GroupVersion info, conversion hubs, controllers
// and the per-group cmd/provider main templates (via the MainTemplate).
//go:generate go run ../cmd/generator ..

// Generate deepcopy methodsets and CRD manifests for both the cluster-scoped
// and namespaced API trees.
//go:generate go run -tags generate sigs.k8s.io/controller-tools/cmd/controller-gen object:headerFile=../hack/boilerplate.go.txt paths=./... crd:crdVersions=v1,allowDangerousTypes=true output:artifacts:config=../package/crds

// Generate crossplane-runtime methodsets (resource.Managed / resource.ManagedList)
// and the cross-resource reference resolvers.
//go:generate go run -tags generate github.com/crossplane/crossplane-tools/cmd/angryjet generate-methodsets --header-file=../hack/boilerplate.go.txt ./...

package apis

import (
	_ "github.com/crossplane/crossplane-tools/cmd/angryjet"     //nolint:typecheck
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"         //nolint:typecheck
)
