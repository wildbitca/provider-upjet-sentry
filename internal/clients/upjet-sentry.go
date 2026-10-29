package clients

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"

	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/upjet/v2/pkg/terraform"

	clusterv1beta1 "github.com/wildbitca/provider-upjet-sentry/apis/cluster/v1beta1"
	namespacedv1beta1 "github.com/wildbitca/provider-upjet-sentry/apis/namespaced/v1beta1"
)

const (
	// error messages
	errNoProviderConfig     = "no providerConfigRef provided"
	errGetProviderConfig    = "cannot get referenced ProviderConfig"
	errTrackUsage           = "cannot track ProviderConfig usage"
	errExtractCredentials   = "cannot extract credentials"
	errUnmarshalCredentials = "cannot unmarshal sentry credentials as JSON"

	errMissingToken   = "sentry credentials must contain a non-empty \"token\""
	errInvalidBaseURL = "sentry credentials \"base_url\" must be an https:// URL with a host, without user information, query or fragment, ending with \"/api/\" (for example https://sentry.io/api/)"

	// Sentry provider config keys
	// https://registry.terraform.io/providers/jianyuan/sentry/latest/docs
	keyToken   = "token"
	keyBaseURL = "base_url"
)

// credentialKeys is the list of Sentry provider config keys to copy from credentials.
var credentialKeys = []string{keyToken, keyBaseURL}

// setConfigFromCreds copies the listed keys from creds into cfg, skipping
// empty values so that the provider's own defaults and environment fallbacks
// (for example the default base_url) are not overridden by an empty string.
func setConfigFromCreds(cfg map[string]any, creds map[string]string, keys []string) {
	for _, k := range keys {
		if v, ok := creds[k]; ok && v != "" {
			cfg[k] = v
		}
	}
}

// parseCredentials decodes the credentials JSON. The error is generic on
// purpose: it must never echo any part of the (secret) input.
func parseCredentials(data []byte) (map[string]string, error) {
	creds := map[string]string{}
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, errors.New(errUnmarshalCredentials)
	}
	return creds, nil
}

// validBaseURL reports whether u is an https URL with a host, no userinfo,
// query or fragment, and a path without dot segments ending in "/api/".
func validBaseURL(u string) bool {
	if !strings.HasPrefix(u, "https://") || !strings.HasSuffix(u, "/api/") {
		return false
	}
	parsed, err := url.Parse(u)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" {
		return false
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
		return false
	}
	return !hasDotSegment(parsed.Path)
}

func hasDotSegment(path string) bool {
	for _, seg := range strings.Split(path, "/") {
		if seg == "." || seg == ".." {
			return true
		}
	}
	return false
}

// buildConfiguration validates the credentials and returns the Terraform
// provider configuration. Errors never include credential values.
func buildConfiguration(raw map[string]string) (map[string]any, error) {
	creds := make(map[string]string, len(raw))
	for k, v := range raw {
		creds[k] = strings.TrimSpace(v)
	}
	if creds[keyToken] == "" {
		return nil, errors.New(errMissingToken)
	}
	// https only: the token travels in every request, so a plain http:// or
	// scheme-less base_url would send it in clear or to an unintended host.
	if u := creds[keyBaseURL]; u != "" && !validBaseURL(u) {
		return nil, errors.New(errInvalidBaseURL)
	}
	cfg := map[string]any{}
	setConfigFromCreds(cfg, creds, credentialKeys)
	return cfg, nil
}

// TerraformSetupBuilder builds Terraform a terraform.SetupFn function which
// returns Terraform provider setup configuration
func TerraformSetupBuilder(version, providerSource, providerVersion string) terraform.SetupFn {
	return func(ctx context.Context, client client.Client, mg resource.Managed) (terraform.Setup, error) {
		ps := terraform.Setup{
			Version: version,
			Requirement: terraform.ProviderRequirement{
				Source:  providerSource,
				Version: providerVersion,
			},
		}

		pcSpec, err := resolveProviderConfig(ctx, client, mg)
		if err != nil {
			return terraform.Setup{}, errors.Wrap(err, "cannot resolve provider config")
		}

		data, err := resource.CommonCredentialExtractor(ctx, pcSpec.Credentials.Source, client, pcSpec.Credentials.CommonCredentialSelectors)
		if err != nil {
			return ps, errors.Wrap(err, errExtractCredentials)
		}
		creds, err := parseCredentials(data)
		if err != nil {
			return ps, err
		}

		cfg, err := buildConfiguration(creds)
		if err != nil {
			return ps, err
		}
		ps.Configuration = cfg
		return ps, nil
	}
}

func toSharedPCSpec(pc *clusterv1beta1.ProviderConfig) (*namespacedv1beta1.ProviderConfigSpec, error) {
	if pc == nil {
		return nil, nil
	}
	data, err := json.Marshal(pc.Spec)
	if err != nil {
		return nil, err
	}

	var mSpec namespacedv1beta1.ProviderConfigSpec
	err = json.Unmarshal(data, &mSpec)
	return &mSpec, err
}

func resolveProviderConfig(ctx context.Context, crClient client.Client, mg resource.Managed) (*namespacedv1beta1.ProviderConfigSpec, error) {
	switch managed := mg.(type) {
	case resource.LegacyManaged: //nolint:staticcheck // cluster-scoped MRs still implement the legacy interface
		return resolveLegacy(ctx, crClient, managed)
	case resource.ModernManaged:
		return resolveModern(ctx, crClient, managed)
	default:
		return nil, errors.New("resource is not a managed resource")
	}
}

func resolveLegacy(ctx context.Context, client client.Client, mg resource.LegacyManaged) (*namespacedv1beta1.ProviderConfigSpec, error) { //nolint:staticcheck // see resolveProviderConfig
	configRef := mg.GetProviderConfigReference()
	if configRef == nil {
		return nil, errors.New(errNoProviderConfig)
	}
	pc := &clusterv1beta1.ProviderConfig{}
	if err := client.Get(ctx, types.NamespacedName{Name: configRef.Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetProviderConfig)
	}

	t := resource.NewLegacyProviderConfigUsageTracker(client, &clusterv1beta1.ProviderConfigUsage{})
	if err := t.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackUsage)
	}

	return toSharedPCSpec(pc)
}

func resolveModern(ctx context.Context, crClient client.Client, mg resource.ModernManaged) (*namespacedv1beta1.ProviderConfigSpec, error) {
	configRef := mg.GetProviderConfigReference()
	if configRef == nil {
		return nil, errors.New(errNoProviderConfig)
	}

	pcRuntimeObj, err := crClient.Scheme().New(namespacedv1beta1.SchemeGroupVersion.WithKind(configRef.Kind))
	if err != nil {
		return nil, errors.Wrap(err, "unknown GVK for ProviderConfig")
	}
	pcObj, ok := pcRuntimeObj.(client.Object)
	if !ok {
		return nil, errors.New("provider config kind is not a Kubernetes object")
	}

	// Namespace will be ignored if the PC is a cluster-scoped type
	if err := crClient.Get(ctx, types.NamespacedName{Name: configRef.Name, Namespace: mg.GetNamespace()}, pcObj); err != nil {
		return nil, errors.Wrap(err, errGetProviderConfig)
	}

	var pcSpec namespacedv1beta1.ProviderConfigSpec
	pcu := &namespacedv1beta1.ProviderConfigUsage{}
	switch pc := pcObj.(type) {
	case *namespacedv1beta1.ProviderConfig:
		pcSpec = pc.Spec
		if pcSpec.Credentials.SecretRef != nil {
			pcSpec.Credentials.SecretRef.Namespace = mg.GetNamespace()
		}
	case *namespacedv1beta1.ClusterProviderConfig:
		pcSpec = pc.Spec
	default:
		return nil, errors.New("unknown provider config type")
	}
	t := resource.NewProviderConfigUsageTracker(crClient, pcu)
	if err := t.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackUsage)
	}
	return &pcSpec, nil
}
