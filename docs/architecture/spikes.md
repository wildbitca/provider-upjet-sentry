# Spikes S0-S4: results and decisions

Run on 2026-10-29 against `jianyuan/sentry` v0.15.8, Upjet fork `github.com/wildbitca/upjet/v2` v2.2.7, OpenTofu 1.11.5 (S1-S5; S6 and the notes marked 2026-10-29b were re-run on upstream `github.com/crossplane/upjet/v2` v2.5.1 and OpenTofu 1.13.2). No Sentry token was available, so every verdict covers generation and static analysis only. Runtime behavior is **unverified**.

| # | Question | Verdict |
|---|---|---|
| S0 | Does Upjet v2.2.7 generate typed CRD fields for Framework `nested_type` attributes? | **GO** |
| S1 | Composite id vs bare id in the synthetic state | **Bare id wins** (plan fallback adopted) |
| S2 | CRD size of `sentry_alert` and `sentry_issue_alert` | **GO** (largest CRD 148 KiB) |
| S3 | Resources without an `id` attribute | **GO** (round-trip covered by unit tests; runtime unverified) |
| S4 | `dynamic` attributes and CEL, `jsontypes.Normalized` | **GO** (no `dynamic` type attribute exists; JSON strings stay plain strings) |

No resource was deferred. All 24 resources are generated in both API trees.

## S0: nested_type attributes

Method: a scratch copy of the repo with only `sentry_team`, `sentry_project`, `sentry_key`, `sentry_cron_monitor` (plus `sentry_organization`, required because references need their target in the include list), then `make generate`.

Result: the generator ran to completion and produced typed objects. `CronMonitor.spec.forProvider.schedule` has `crontab`, `intervalUnit`, `intervalValue`; `owner` has `teamId` (with `teamIdRef`/`teamIdSelector`) and `userId`; `Project.spec.forProvider.clientSecurity` has `allowedDomains`, `scrapeJavascript`, `securityTokenHeader`, `verifyTlsSsl` and a `securityTokenSecretRef`. The fork needed no patch.

## S1: composite id vs bare id

Finding from `upjet/pkg/terraform/files.go` (`EnsureTFState`): before every refresh Upjet writes a synthetic state whose `id` is the output of `GetIDFn`, plus all parameters. For Framework resources the provider's `Read` uses the state `id`, so `GetIDFn` must reproduce exactly the `id` that the provider itself stores in state. The upstream research table of stored state ids therefore decides the strategy:

- Bare id in state (key, issue_alert, alert, monitors, notification_action, organization_repository, symbol_source, opsgenie, pagerduty, code_mapping, plugin): `IdentifierFromProvider`. The `organization` and `project` attributes that `Read` also needs are copied from the parameters into the synthetic state.
- Composite id in state (organization_member, metric_alert, dashboard): `TemplatedStringAsIdentifier` with the same composite.
- Slug is the id (organization, team, project): custom config that seeds `slug` from the external name.

The composite-`GetIDFn` variant of the plan table was dropped for the bare-id resources because writing `org/project/id` into the state `id` risks `Read` requesting the wrong key. This is reasoned from the code, not observed against a live API.

## S2: CRD size

`make generate` takes about 10 seconds for 24 resources x 2 scopes. 53 CRDs exist (48 resources + 5 ProviderConfig kinds). The largest are `Alert` (148 KiB), `IssueAlert` (96 KiB) and `CronMonitor` (90 KiB). All are far below the 700 KiB target and the 1 MiB etcd limit, so no deprecated attribute had to be trimmed for size. Deprecated attributes were still omitted for correctness (see `external-names.md` and `resource-matrix.md`).

## S3: id-less resources

`sentry_project_ownership` and `sentry_all_projects_spike_protection` have no `id` attribute. Upjet detects this from the schema (`WithHasIDAttribute`) and does not write `id` into the synthetic state. Identity is built from parameters: `org/project` and `org`. Upstream v2.5.1 also writes no id for them (`if fp.hasTFID` in `files.go`). The fork-only empty-id guard and `patches/apply.sh` were removed with the fork; see S6 for what that changes for the other resources.

`config/external_name_test.go` covers `GetIDFn`, `GetExternalNameFn` and the missing-parameter errors for both. A real create/observe/delete cycle is **unverified**.

## S4: dynamic attributes and JSON strings

The v0.15.8 schema has no attribute of type `dynamic` (one grep hit was the word in a description). The CEL "required parameter" rules Upjet emits are on plain attributes. Offline CRD validation is covered by the CRD smoke test in `config/crd_smoke_test.go`; server-side CRD admission was not run because no cluster is in scope.

`uptime_monitor.assertion_json` is a string whose upstream type is `jsontypes.Normalized`. Semantic equality is implemented inside the provider, so a perpetual diff is unlikely, but it is **unverified** at runtime; the mitigation if it appears is a custom `TerraformConversion` (cloudflare `3329922` precedent).

## S6 - empty id in the first synthetic state on upstream upjet v2.5.1 (2026-10-29b)

Setup: OpenTofu 1.13.2, `jianyuan/sentry` v0.15.8, a local HTTP server that answers 200 on `/api/0/internal/health/` and 404 elsewhere, and a state file holding `sentry_key.k` with `id = ""` (what upjet v2.5.1 writes before the first Create of an `IdentifierFromProvider` resource, because the `if fp.hasTFID` branch has no empty-id guard; crossplane/upjet#789 is open and the fork tag `v2.5.1-wb.1` does not carry the guard either).

Result: `tofu apply` refreshes the resource first and the provider calls `GET /api/0/projects/acme/web/keys//`, then fails with `Not found / No matching client key found`. The create step is never reached. The upjet CLI runtime runs `apply` with refresh on, so the same failure is expected on the first reconcile of the 12 `IdentifierFromProvider` resources against a real Sentry server (whose answer to `keys//` may differ, so this is **unverified live**).

Consequence: the guard that lived in the fork (`fp.hasTFID && tfID != ""`) is the open risk of running on upstream only. Options, in order of preference: (1) land crossplane/upjet#789 upstream and bump; (2) validate live (create a throwaway `ProjectKey`) and, if it fails, re-add a one-line fork `replace`; (3) there is no in-repo workaround, since `WorkspaceStore` is a concrete type and `GetIDFn` cannot suppress the `id` write.

Resolution (2026-10-29c): option (2) was taken in advance: `go.mod` pins `github.com/wildbitca/upjet/v2` v2.5.1-wb.1, which carries the guard (see [upjet-fork.md](upjet-fork.md)).
