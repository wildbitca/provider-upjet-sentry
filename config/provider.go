package config

import (
	// Note(turkenh): we are importing this to embed provider schema document
	_ "embed"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"

	"github.com/wildbitca/provider-upjet-sentry/hack"

	alertCluster "github.com/wildbitca/provider-upjet-sentry/config/cluster/alert"
	dashboardCluster "github.com/wildbitca/provider-upjet-sentry/config/cluster/dashboard"
	integrationCluster "github.com/wildbitca/provider-upjet-sentry/config/cluster/integration"
	monitorCluster "github.com/wildbitca/provider-upjet-sentry/config/cluster/monitor"
	organizationCluster "github.com/wildbitca/provider-upjet-sentry/config/cluster/organization"
	projectCluster "github.com/wildbitca/provider-upjet-sentry/config/cluster/project"
	teamCluster "github.com/wildbitca/provider-upjet-sentry/config/cluster/team"
	alertNamespaced "github.com/wildbitca/provider-upjet-sentry/config/namespaced/alert"
	dashboardNamespaced "github.com/wildbitca/provider-upjet-sentry/config/namespaced/dashboard"
	integrationNamespaced "github.com/wildbitca/provider-upjet-sentry/config/namespaced/integration"
	monitorNamespaced "github.com/wildbitca/provider-upjet-sentry/config/namespaced/monitor"
	organizationNamespaced "github.com/wildbitca/provider-upjet-sentry/config/namespaced/organization"
	projectNamespaced "github.com/wildbitca/provider-upjet-sentry/config/namespaced/project"
	teamNamespaced "github.com/wildbitca/provider-upjet-sentry/config/namespaced/team"
)

const (
	resourcePrefix = "upjet-sentry"
	modulePath     = "github.com/wildbitca/provider-upjet-sentry"
)

//go:embed schema.json
var providerSchema string

//go:embed provider-metadata.yaml
var providerMetadata string

// GetProvider returns provider configuration
func GetProvider() *ujconfig.Provider {
	pc := ujconfig.NewProvider([]byte(providerSchema), resourcePrefix, modulePath, []byte(providerMetadata),
		ujconfig.WithRootGroup("upjet-sentry.upbound.io"),
		ujconfig.WithIncludeList(ExternalNameConfigured()),
		ujconfig.WithFeaturesPackage("internal/features"),
		ujconfig.WithMainTemplate(hack.MainTemplate),
		ujconfig.WithDefaultResourceOptions(
			ExternalNameConfigurations(),
		))

	for _, configure := range []func(provider *ujconfig.Provider){
		organizationCluster.Configure,
		teamCluster.Configure,
		projectCluster.Configure,
		alertCluster.Configure,
		monitorCluster.Configure,
		integrationCluster.Configure,
		dashboardCluster.Configure,
	} {
		configure(pc)
	}

	pc.ConfigureResources()
	return pc
}

// GetProviderNamespaced returns the namespaced provider configuration
func GetProviderNamespaced() *ujconfig.Provider {
	pc := ujconfig.NewProvider([]byte(providerSchema), resourcePrefix, modulePath, []byte(providerMetadata),
		ujconfig.WithRootGroup("upjet-sentry.m.upbound.io"),
		ujconfig.WithIncludeList(ExternalNameConfigured()),
		ujconfig.WithFeaturesPackage("internal/features"),
		ujconfig.WithMainTemplate(hack.MainTemplate),
		ujconfig.WithDefaultResourceOptions(
			ExternalNameConfigurations(),
		),
		ujconfig.WithExampleManifestConfiguration(ujconfig.ExampleManifestConfiguration{
			ManagedResourceNamespace: "crossplane-system",
		}))

	for _, configure := range []func(provider *ujconfig.Provider){
		organizationNamespaced.Configure,
		teamNamespaced.Configure,
		projectNamespaced.Configure,
		alertNamespaced.Configure,
		monitorNamespaced.Configure,
		integrationNamespaced.Configure,
		dashboardNamespaced.Configure,
	} {
		configure(pc)
	}

	pc.ConfigureResources()
	return pc
}
