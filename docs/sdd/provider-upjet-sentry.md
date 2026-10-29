# SDD: provider-upjet-sentry

Status: implemented offline (generation, build, lint, unit tests, package build). Not published. Live behavior unverified.

## 1. Goal

A public Crossplane provider, `github.com/wildbitca/provider-upjet-sentry` (Apache-2.0), that manages Sentry configuration as Kubernetes resources by wrapping `jianyuan/sentry` v0.15.8 through OpenTofu. Same shape as the sibling `provider-upjet-*` repositories.

## 2. Scope

In scope:

- All 24 managed resources of the pinned upstream release, in a cluster-scoped and a namespaced API tree (48 CRDs) plus `ProviderConfig`, `ProviderConfigUsage` and (namespaced) `ClusterProviderConfig`.
- Seven API groups and seven sub-providers plus a family package.
- Credentials as one JSON Secret key (`token`, `base_url`), skipping empty values.
- Examples, docs, CI and publish workflows, a package build dry-run.

Non-goals:

- Data sources (20 upstream) and provider functions.
- SDK or Framework in-process (no-fork) controllers.
- Tagging, publishing, repository creation, secrets, cluster installation (separate approval).
- Multi-arch images (linux/amd64 only, as the sibling repos).
- Live verification against Sentry.

## 3. Group contract

| Group | Kinds |
|---|---|
| `organization` | `Organization`, `OrganizationMember`, `OrganizationCodeMapping`, `OrganizationRepository` |
| `team` | `Team`, `TeamMember` |
| `project` | `Project`, `ProjectKey`, `ProjectPlugin`, `ProjectOwnership`, `ProjectSpikeProtection`, `AllProjectsSpikeProtection`, `ProjectInboundDataFilter`, `ProjectSymbolSource` |
| `alert` | `IssueAlert`, `MetricAlert`, `Alert`, `NotificationAction` |
| `monitor` | `CronMonitor`, `MetricMonitor`, `UptimeMonitor` |
| `integration` | `IntegrationOpsgenie`, `IntegrationPagerduty` |
| `dashboard` | `Dashboard` |

API groups are `<group>.upjet-sentry.upbound.io` and `<group>.upjet-sentry.m.upbound.io`, version `v1alpha1`. Names are never changed after release; resources never move between groups. Reasoning for the split: publish time scales with the group count (cloudflare: 66 groups = 28-68 minutes), while `alert` and `organization` images stay small.

## 4. Requirements and acceptance criteria

| # | Requirement | Evidence |
|---|---|---|
| R1 | 24 resources generate in both trees | `TestGroupAndKindContract`, `TestCRDsPerGroup` (48 resource CRDs) |
| R2 | Group and Kind names match section 3 | `config/provider_test.go` |
| R3 | External names per `docs/architecture/external-names.md` | `config/external_name_test.go` |
| R4 | Credentials skip empty values, require `token`, validate `base_url`, never leak values | `internal/clients/upjet-sentry_test.go` |
| R5 | Sensitive inputs are secret references; deprecated and secret-bearing outputs omitted | `TestSensitiveInputsAreNotPlainParameters`, `TestNoPlainSecretsInCRDs`, `TestDeprecatedAttributesAreOmitted` |
| R6 | Every CRD < 700 KiB | `loadCRDs` budget (largest 148 KiB) |
| R7 | One valid example per resource | `TestExamplesValidateAgainstCRDs` |
| R8 | Group list identical everywhere | `scripts/check-groups.sh` (CI `guardrails`) |
| R9 | No piped installers, actions pinned | `scripts/check-workflows.sh` |
| R10 | `make generate`, `build`, `lint`, `test` exit 0; `generate` idempotent | command evidence in `docs/architecture/test-report.md` |
| R11 | Package build produces 8 xpkgs, each with only its group's CRDs | `docs/architecture/test-report.md` |

## 5. Known limits

See the README "Known issues" list and `docs/architecture/spikes.md`. Everything about the real API (token scopes, import flows, deletion semantics of singleton-like resources, the effect of the per-configure health check) is unverified.

## 6. Decisions

- CLI (OpenTofu) mode only, `IncludeList` only. `internal/` Framework packages cannot be imported (see `overview.md`).
- Upstream `github.com/crossplane/upjet/v2` v2.5.1 (no fork; the fork-only CEL/dynamic fixes are not needed by this schema, and the empty-id guard is an open risk, see `docs/architecture/spikes.md` S6), `crossplane-runtime/v2` v2.4.0, `crossplane/apis/v2` v2.4.1, OpenTofu 1.13.2.
- Bare-id synthetic state for provider-assigned ids (spike S1).
- Explicit Kind for every resource (Upjet's default drops the first word after the prefix).
- `OrganizationCodeMapping.project_id` is not a reference (import cycle between groups).
