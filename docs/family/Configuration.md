# Installing the provider family

Two ways to install: the family (recommended) or a single sub-provider that pulls the family in as a dependency.

## Family and sub-providers

`provider-family-sentry` holds the `ProviderConfig` CRDs and the shared runtime. Each sub-provider holds the CRDs of one API group and depends on the family (`>= <version>`).

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-sentry-team
spec:
  package: xpkg.upbound.io/wildbitca/provider-sentry-team:v1.0.0
```

Installing a sub-provider installs `provider-family-sentry` automatically. Install only the groups you use: `organization`, `team`, `project`, `alert`, `monitor`, `integration`, `dashboard`.

## Monolith

The repository also builds a monolith provider (`provider-upjet-sentry`, all 24 resources, 48 CRDs) for local testing. It is not published.

## ProviderConfig

See the README and `docs/architecture/credentials.md`. One `ProviderConfig` per Sentry organization is recommended so that a token's scopes match the resources it manages.

## Namespaced API

The `*.m.upbound.io` groups are namespaced managed resources (Crossplane 2). They use a namespaced `ProviderConfig` or a `ClusterProviderConfig`.
