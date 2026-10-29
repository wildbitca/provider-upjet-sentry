# Provider Sentry

`provider-sentry` is a [Crossplane](https://crossplane.io/) provider built with [Upjet](https://github.com/crossplane/upjet). It exposes XRM-conformant managed resources for [Sentry](https://sentry.io) by wrapping the [`jianyuan/sentry`](https://registry.terraform.io/providers/jianyuan/sentry/latest/docs) Terraform provider (v0.15.8), run through OpenTofu.

It covers all 24 managed resources of the upstream provider, in a cluster-scoped and a namespaced API tree.

> **Status: first stable release (v1.0.0).** Everything was verified offline (code generation, build, lint, unit tests, package build). Runtime behavior against a live Sentry organization is **unverified**: see [Known issues](#known-issues).

## Getting started

This provider uses the **family** distribution model: install only the sub-providers you need. The first sub-provider you install pulls in `provider-family-sentry`, which owns the shared `ProviderConfig`. See [`docs/family/Configuration.md`](docs/family/Configuration.md).

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-sentry-project
spec:
  package: xpkg.upbound.io/wildbitca/provider-sentry-project:v1.0.0
```

### Sub-providers

| Sub-provider | API group (cluster / namespaced) | Kinds |
|---|---|---|
| `provider-sentry-organization` | `organization.upjet-sentry.upbound.io` / `.m.upbound.io` | `Organization`, `OrganizationMember`, `OrganizationCodeMapping`, `OrganizationRepository` |
| `provider-sentry-team` | `team.upjet-sentry...` | `Team`, `TeamMember` |
| `provider-sentry-project` | `project.upjet-sentry...` | `Project`, `ProjectKey`, `ProjectPlugin`, `ProjectOwnership`, `ProjectSpikeProtection`, `AllProjectsSpikeProtection`, `ProjectInboundDataFilter`, `ProjectSymbolSource` |
| `provider-sentry-alert` | `alert.upjet-sentry...` | `IssueAlert`, `MetricAlert`, `Alert`, `NotificationAction` |
| `provider-sentry-monitor` | `monitor.upjet-sentry...` | `CronMonitor`, `MetricMonitor`, `UptimeMonitor` |
| `provider-sentry-integration` | `integration.upjet-sentry...` | `IntegrationOpsgenie`, `IntegrationPagerduty` |
| `provider-sentry-dashboard` | `dashboard.upjet-sentry...` | `Dashboard` |

The group and Kind names are a contract: they are never renamed after release. The full mapping is in [`docs/architecture/resource-matrix.md`](docs/architecture/resource-matrix.md).

### ProviderConfig

Credentials live in a single Secret key holding a JSON object:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: sentry-creds
  namespace: crossplane-system
type: Opaque
stringData:
  credentials: |
    {
      "token": "REPLACE_WITH_SENTRY_AUTH_TOKEN",
      "base_url": "https://sentry.io/api/"
    }
---
apiVersion: upjet-sentry.upbound.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      name: sentry-creds
      namespace: crossplane-system
      key: credentials
```

- `token` is required: an Organization Auth Token or an Internal Integration token.
- `base_url` is optional and defaults to `https://sentry.io/api/`. It must be an `https://` URL ending with `/api/`. Use `https://us.sentry.io/api/` or `https://de.sentry.io/api/` for regional SaaS, or your own host for self-hosted Sentry.
- Empty values are skipped, so the provider's own defaults apply.

Details, token scopes and the namespaced `ProviderConfig` / `ClusterProviderConfig` variants: [`docs/architecture/credentials.md`](docs/architecture/credentials.md).

### Example

```yaml
apiVersion: team.upjet-sentry.upbound.io/v1alpha1
kind: Team
metadata:
  name: backend
spec:
  forProvider:
    organization: my-org
    name: Backend
  providerConfigRef:
    name: default
```

One manifest per resource lives in [`examples/`](examples).

### Rate limiting

Sentry rate limits API calls, and the provider performs a health check on every configure. For large deployments throttle the reconcile rate with a `DeploymentRuntimeConfig`:

```yaml
apiVersion: pkg.crossplane.io/v1beta1
kind: DeploymentRuntimeConfig
metadata:
  name: sentry-throttled
spec:
  deploymentTemplate:
    spec:
      selector: {}
      template:
        spec:
          containers:
            - name: package-runtime
              args:
                - --poll=30m
```

## Known issues

These come from the upstream provider's issue tracker and from offline analysis. None of them could be reproduced here because no Sentry token was available.

- Upstream "inconsistent result after apply" reports (#938, #814, #843, #694, #942) will surface as reconcile errors on the affected resources.
- `NotificationAction`: orphaned actions after failed creates (#937, #960). Manage with care; creates are not retried automatically.
- `Project`: changing `slug` is destructive (#725). The slug is the external name and must be treated as immutable.
- `TeamMember`: import can hit a nil-pointer (#651); prefer creating over observing.
- `MetricMonitor`: import omits `project` (#964).
- `ProjectOwnership` and `AllProjectsSpikeProtection` have no `id` in the upstream schema; their identity is built from parameters. Round-trip is covered by unit tests only.
- `Organization` creation is generally not possible on Sentry SaaS; use it in observe-only mode or on self-hosted.
- `IssueAlert` is on schema version 2; test against fresh creates, not upgraded state.

## Development

```bash
make generate     # schema, CRDs, controllers
make build
make lint
make test
scripts/check-groups.sh      # the group list must match everywhere
scripts/check-workflows.sh   # no piped installers, actions pinned to SHAs
```

See [`CLAUDE.md`](CLAUDE.md) for repository conventions and [`docs/`](docs) for the design.

## License

Apache-2.0, see [`LICENSE`](LICENSE) and [`NOTICE`](NOTICE). The wrapped `jianyuan/terraform-provider-sentry` is MIT licensed; its license text ships inside the provider images.
