# Provider Sentry

`provider-sentry` is a [Crossplane](https://crossplane.io/) provider built using [Upjet](https://github.com/crossplane/upjet) that exposes XRM-conformant managed resources for [Sentry](https://docs.sentry.io/api/), wrapping the [`jianyuan/sentry`](https://registry.terraform.io/providers/jianyuan/sentry/latest/docs) Terraform provider.

This provider is distributed as a **provider family**. Install only the sub-provider you need; it automatically pulls in the shared `ProviderConfig` via `provider-family-sentry`.

| Package | Resources |
|---------|-----------|
| `provider-family-sentry` | ProviderConfig kinds (auto-installed) |
| `provider-sentry-organization` | Organization, OrganizationMember, OrganizationCodeMapping, OrganizationRepository |
| `provider-sentry-team` | Team, TeamMember |
| `provider-sentry-project` | Project, ProjectKey, ProjectPlugin, ProjectOwnership, ProjectSpikeProtection, AllProjectsSpikeProtection, ProjectInboundDataFilter, ProjectSymbolSource |
| `provider-sentry-alert` | IssueAlert, MetricAlert, Alert, NotificationAction |
| `provider-sentry-monitor` | CronMonitor, MetricMonitor, UptimeMonitor |
| `provider-sentry-integration` | IntegrationOpsgenie, IntegrationPagerduty |
| `provider-sentry-dashboard` | Dashboard |

## Install

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-sentry-team
spec:
  package: xpkg.upbound.io/wildbitca/provider-sentry-team:v0.1.0
```

## ProviderConfig

`token` is an Organization Auth Token or an Internal Integration token. `base_url` is optional (default `https://sentry.io/api/`) and must be an `https://` URL ending with `/api/`.

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

## Report a Bug

Open an [issue](https://github.com/wildbitca/provider-upjet-sentry/issues).
