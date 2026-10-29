# Changelog

All notable changes to this provider family are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2026-10-29

First release. Wraps `jianyuan/sentry` v0.15.8 (OpenTofu 1.13.2) on Upjet v2.5.1 pinned to the `wildbitca/upjet` fork tag `v2.5.1-wb.1` (see `docs/architecture/upjet-fork.md`). The `0.1.0` entry below was never tagged or published; `1.0.0` supersedes it and is the first published version.

### Breaking changes

- Relative to the unpublished pre-release builds only (nothing was ever published): ProviderConfig API types moved to `apis/v2/core/v2` of `github.com/crossplane/crossplane/apis/v2`; conversion webhooks are registered when a certificate directory is configured; CRDs gain the additive `status.lastHandledReconcileAt` field. API groups and Kinds are the public contract from here on and will not be renamed within 1.x.
- Migration note: there is nothing to migrate from a published release. Install `xpkg.upbound.io/wildbitca/provider-family-sentry:v1.0.0` or the per-group `provider-sentry-<group>:v1.0.0` packages. Anyone who built a pre-release from source must delete and reinstall its CRDs, and requires Crossplane v2.

### Changed

- Upjet is `github.com/crossplane/upjet/v2` v2.5.1 with a `replace` to `github.com/wildbitca/upjet/v2` v2.5.1-wb.1 (numeric state ids, no CEL for required dynamic parameters, empty-id guard). `crossplane-runtime/v2` v2.4.0, direct dependency `github.com/crossplane/crossplane/apis/v2` v2.4.1, Go 1.26.8, controller-gen v0.20, controller-runtime v0.25, client-go v0.37, terraform-plugin-sdk v2.40.1, terraform-plugin-framework v1.19.0, grpc v1.85.0.
- OpenTofu 1.11.5 -> 1.13.2 (Makefile, publish workflow, docs). `jianyuan/sentry` stays at v0.15.8, the latest release.
- The provider mains register the generated conversion webhooks (`SetupWebhookWithManager_<group>`) when a certificate directory is configured.
- Package icon is the official Sentry glyph (source and trademark notes in `docs/architecture/release-and-security.md`).

### Fixed

- Observe-only and adoption imports no longer fail for resources whose Terraform importer needs a composite id while the state id is bare (`Team`, `Project`, `ProjectKey`, `IssueAlert`, `Alert`, `CronMonitor`, `MetricMonitor`, `UptimeMonitor`, `NotificationAction`, `OrganizationCodeMapping`, `OrganizationRepository`, `ProjectSymbolSource`, `IntegrationOpsgenie`, `IntegrationPagerduty`, `ProjectPlugin`). `ExternalName.GetImportIDFn` builds the composite string; `GetIDFn` stays bare.
- First reconcile of `IdentifierFromProvider` resources no longer refreshes an empty id (fork empty-id guard, `docs/architecture/spikes.md` S6).

## [0.1.0] - never released

Internal pre-release snapshot, folded into 1.0.0. Wrapped `jianyuan/sentry` v0.15.8 (OpenTofu 1.11.5) with the Upjet fork `wildbitca/upjet` v2.2.7.

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
