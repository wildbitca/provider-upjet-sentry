# Changelog

All notable changes to this provider family are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - unreleased

First version. Wraps `jianyuan/sentry` v0.15.8 (OpenTofu 1.11.5) with the Upjet fork `wildbitca/upjet` v2.2.7. Not tagged or published yet.

### Added

- All 24 managed resources of the upstream provider in seven API groups, each in a cluster-scoped (`<group>.upjet-sentry.upbound.io`) and a namespaced (`<group>.upjet-sentry.m.upbound.io`) API:
  - `organization`: `Organization`, `OrganizationMember`, `OrganizationCodeMapping`, `OrganizationRepository`
  - `team`: `Team`, `TeamMember`
  - `project`: `Project`, `ProjectKey`, `ProjectPlugin`, `ProjectOwnership`, `ProjectSpikeProtection`, `AllProjectsSpikeProtection`, `ProjectInboundDataFilter`, `ProjectSymbolSource`
  - `alert`: `IssueAlert`, `MetricAlert`, `Alert`, `NotificationAction`
  - `monitor`: `CronMonitor`, `MetricMonitor`, `UptimeMonitor`
  - `integration`: `IntegrationOpsgenie`, `IntegrationPagerduty`
  - `dashboard`: `Dashboard`
- `ProviderConfig` (cluster and namespaced), `ClusterProviderConfig` (namespaced) and `ProviderConfigUsage`.
- Credentials as one JSON Secret key with `token` and optional `base_url`; empty values are skipped, a missing token and a `base_url` without the `/api/` suffix are rejected.
- Secret references for sensitive inputs (symbol source keys, Opsgenie and PagerDuty integration keys, project security token); `ProjectKey` publishes `secret` and `dsn` as connection details.
- Per-resource examples under `examples/`, offline validated against the generated CRDs.
- Family packaging: `provider-family-sentry` plus one `provider-sentry-<group>` per group.
- CI guardrails: consistent group lists, no piped installers, SHA-pinned actions, no expressions inside `run:` scripts, no template residue, a pinned gitleaks scan over the full history.

### Changed relative to the sibling providers' pipeline

- The Crossplane CLI in `publish.yml` is a pinned release with a verified SHA-256 instead of `curl | sh` from `master`.
- Provider and OpenTofu archives and the upstream license in the image are verified against SHA-256 values pinned in the repository; the base image is pinned by digest; the upstream MIT license is shipped in the image.
- Publish jobs use no caches; the release version is validated as semver; `Tag` only runs from `main`.
- Workflow secrets are passed through `env:`; default token permissions are read-only.

### Omitted upstream attributes

- `Team.team_id`, `ProjectKey.dsn_public` / `dsn_csp` / `dsn_secret`, `IssueAlert.actions` / `conditions` / `filters` and `Alert.action_filters.actions.plugin` are deprecated upstream and not exposed.

### Known limitations

- Runtime behavior against a live Sentry organization is unverified.
- See the README for the upstream issues that surface as reconcile errors.
