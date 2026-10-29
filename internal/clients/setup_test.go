package clients

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"

	clusterteam "github.com/wildbitca/provider-upjet-sentry/apis/cluster/team/v1alpha1"
	clusterv1beta1 "github.com/wildbitca/provider-upjet-sentry/apis/cluster/v1beta1"
	namespacedteam "github.com/wildbitca/provider-upjet-sentry/apis/namespaced/team/v1alpha1"
	namespacedv1beta1 "github.com/wildbitca/provider-upjet-sentry/apis/namespaced/v1beta1"
)

const (
	testNamespace = "team-a"
	testSecret    = "sentry-creds"
	testKey       = "credentials"
)

func newScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	for _, add := range []func(*runtime.Scheme) error{
		corev1.AddToScheme,
		clusterv1beta1.SchemeBuilder.AddToScheme,
		namespacedv1beta1.SchemeBuilder.AddToScheme,
	} {
		if err := add(s); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

// clusterScopedGet mimics the API server for cluster-scoped kinds: the
// resolver passes the managed resource namespace along and the server ignores it.
type clusterScopedGet struct{ client.Client }

func (c clusterScopedGet) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if _, ok := obj.(*namespacedv1beta1.ClusterProviderConfig); ok {
		key.Namespace = ""
	}
	return c.Client.Get(ctx, key, obj, opts...)
}

func credsSecret(namespace string, data map[string][]byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: testSecret, Namespace: namespace},
		Data:       data,
	}
}

func secretSelector(namespace string) xpv1.CommonCredentialSelectors {
	return xpv1.CommonCredentialSelectors{SecretRef: &xpv1.SecretKeySelector{
		SecretReference: xpv1.SecretReference{Name: testSecret, Namespace: namespace},
		Key:             testKey,
	}}
}

func clusterPC(sel xpv1.CommonCredentialSelectors) *clusterv1beta1.ProviderConfig {
	return &clusterv1beta1.ProviderConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "default", UID: types.UID("pc-uid")},
		Spec: clusterv1beta1.ProviderConfigSpec{Credentials: clusterv1beta1.ProviderCredentials{
			Source: xpv1.CredentialsSourceSecret, CommonCredentialSelectors: sel,
		}},
	}
}

func clusterTeam(ref *xpv1.Reference) *clusterteam.Team {
	t := &clusterteam.Team{ObjectMeta: metav1.ObjectMeta{Name: "t", UID: types.UID("mg-uid")}}
	t.SetProviderConfigReference(ref)
	return t
}

func namespacedTeamMR(ref *xpv1.ProviderConfigReference) *namespacedteam.Team {
	t := &namespacedteam.Team{ObjectMeta: metav1.ObjectMeta{Name: "t", Namespace: testNamespace, UID: types.UID("mg-uid")}}
	t.SetProviderConfigReference(ref)
	return t
}

func TestTerraformSetupBuilder(t *testing.T) {
	good := []byte(`{"token":"` + canary + `","base_url":"https://de.sentry.io/api/"}`)
	wantGood := map[string]any{"token": canary, "base_url": "https://de.sentry.io/api/"}

	cases := map[string]struct {
		objs    []client.Object
		mg      resource.Managed
		want    map[string]any
		wantErr string
	}{
		"ClusterScopedSecretOK": {
			objs: []client.Object{clusterPC(secretSelector("crossplane-system")), credsSecret("crossplane-system", map[string][]byte{testKey: good})},
			mg:   clusterTeam(&xpv1.Reference{Name: "default"}),
			want: wantGood,
		},
		"NamespacedProviderConfigUsesTheManagedResourceNamespace": {
			objs: []client.Object{
				&namespacedv1beta1.ProviderConfig{
					ObjectMeta: metav1.ObjectMeta{Name: "default", Namespace: testNamespace, UID: types.UID("pc-uid")},
					Spec: namespacedv1beta1.ProviderConfigSpec{Credentials: namespacedv1beta1.ProviderCredentials{
						Source: xpv1.CredentialsSourceSecret, CommonCredentialSelectors: secretSelector("ignored"),
					}},
				},
				credsSecret(testNamespace, map[string][]byte{testKey: good}),
			},
			mg:   namespacedTeamMR(&xpv1.ProviderConfigReference{Name: "default", Kind: "ProviderConfig"}),
			want: wantGood,
		},
		"ClusterProviderConfigKeepsItsSecretNamespace": {
			objs: []client.Object{
				&namespacedv1beta1.ClusterProviderConfig{
					ObjectMeta: metav1.ObjectMeta{Name: "shared", UID: types.UID("pc-uid")},
					Spec: namespacedv1beta1.ProviderConfigSpec{Credentials: namespacedv1beta1.ProviderCredentials{
						Source: xpv1.CredentialsSourceSecret, CommonCredentialSelectors: secretSelector("crossplane-system"),
					}},
				},
				credsSecret("crossplane-system", map[string][]byte{testKey: good}),
			},
			mg:   namespacedTeamMR(&xpv1.ProviderConfigReference{Name: "shared", Kind: "ClusterProviderConfig"}),
			want: wantGood,
		},
		"NoProviderConfigReference": {
			mg:      clusterTeam(nil),
			wantErr: errNoProviderConfig,
		},
		"NoProviderConfigReferenceNamespaced": {
			mg:      namespacedTeamMR(nil),
			wantErr: errNoProviderConfig,
		},
		"ProviderConfigNotFound": {
			mg:      clusterTeam(&xpv1.Reference{Name: "missing"}),
			wantErr: errGetProviderConfig,
		},
		"NamespacedProviderConfigNotFound": {
			mg:      namespacedTeamMR(&xpv1.ProviderConfigReference{Name: "missing", Kind: "ProviderConfig"}),
			wantErr: errGetProviderConfig,
		},
		"UnknownProviderConfigKind": {
			mg:      namespacedTeamMR(&xpv1.ProviderConfigReference{Name: "default", Kind: "NoSuchKind"}),
			wantErr: "unknown GVK for ProviderConfig",
		},
		"SecretMissing": {
			objs:    []client.Object{clusterPC(secretSelector("crossplane-system"))},
			mg:      clusterTeam(&xpv1.Reference{Name: "default"}),
			wantErr: errExtractCredentials,
		},
		"SecretKeyMissing": {
			objs:    []client.Object{clusterPC(secretSelector("crossplane-system")), credsSecret("crossplane-system", map[string][]byte{"other": good})},
			mg:      clusterTeam(&xpv1.Reference{Name: "default"}),
			wantErr: errUnmarshalCredentials,
		},
		"SecretValueEmpty": {
			objs:    []client.Object{clusterPC(secretSelector("crossplane-system")), credsSecret("crossplane-system", map[string][]byte{testKey: {}})},
			mg:      clusterTeam(&xpv1.Reference{Name: "default"}),
			wantErr: errUnmarshalCredentials,
		},
		"SecretMalformedJSON": {
			objs:    []client.Object{clusterPC(secretSelector("crossplane-system")), credsSecret("crossplane-system", map[string][]byte{testKey: []byte(`{"token":"` + canary)})},
			mg:      clusterTeam(&xpv1.Reference{Name: "default"}),
			wantErr: errUnmarshalCredentials,
		},
		"SecretNotAJSONObject": {
			objs:    []client.Object{clusterPC(secretSelector("crossplane-system")), credsSecret("crossplane-system", map[string][]byte{testKey: []byte(`"` + canary + `"`)})},
			mg:      clusterTeam(&xpv1.Reference{Name: "default"}),
			wantErr: errUnmarshalCredentials,
		},
		"TokenMissingInJSON": {
			objs:    []client.Object{clusterPC(secretSelector("crossplane-system")), credsSecret("crossplane-system", map[string][]byte{testKey: []byte(`{"base_url":"https://sentry.io/api/"}`)})},
			mg:      clusterTeam(&xpv1.Reference{Name: "default"}),
			wantErr: errMissingToken,
		},
		"InvalidBaseURLDoesNotEchoTheToken": {
			objs:    []client.Object{clusterPC(secretSelector("crossplane-system")), credsSecret("crossplane-system", map[string][]byte{testKey: []byte(`{"token":"` + canary + `","base_url":"http://x/api/"}`)})},
			mg:      clusterTeam(&xpv1.Reference{Name: "default"}),
			wantErr: errInvalidBaseURL,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			kube := fake.NewClientBuilder().WithScheme(newScheme(t)).WithObjects(tc.objs...).Build()
			setup, err := TerraformSetupBuilder("1.11.5", "jianyuan/sentry", "0.15.8")(context.Background(), clusterScopedGet{kube}, tc.mg)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("want error containing %q, got %v", tc.wantErr, err)
				}
				if strings.Contains(err.Error(), canary) {
					t.Fatalf("error leaks the credentials: %v", err)
				}
				if setup.Configuration != nil {
					t.Errorf("configuration must stay empty on error, got %v", setup.Configuration)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if setup.Version != "1.11.5" || setup.Requirement.Source != "jianyuan/sentry" || setup.Requirement.Version != "0.15.8" {
				t.Errorf("wrong version or requirement: %+v", setup)
			}
			cfg := setup.Configuration
			if len(cfg) != len(tc.want) {
				t.Fatalf("want %v, got %v", tc.want, setup.Configuration)
			}
			for k, v := range tc.want {
				if cfg[k] != v {
					t.Errorf("config[%q] = %v, want %v", k, cfg[k], v)
				}
			}
		})
	}
}
