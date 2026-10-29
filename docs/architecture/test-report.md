# Test and verification report

Date 2026-10-29. Host: Linux x64, Go 1.26.1 toolchain (`GOTOOLCHAIN=go1.26.1`, the CI line 1.26; the host default is Go 1.27.3, with which golangci-lint v2.11.4 panics), OpenTofu 1.11.5, Docker 29, golangci-lint v2.11.4 (installed by `build/makelib`). All commands ran locally; nothing was pushed, tagged or published and no cluster was touched.

Written by the implementer role from command output. It is evidence for the verifier (T18), not a sign-off.

## Commands

| Command | Result |
|---|---|
| `make generate` (cold, 24 resources x 2 scopes) | exit 0, about 10 s; 53 CRDs |
| `make generate` a second time | exit 0, `git status` empty (idempotent, equivalent to `make check-diff`) |
| `make vendor vendor.check` | exit 0 (modules tidy, verified) |
| `make build` | exit 0 (binaries `generator`, `config`) |
| `make lint` (golangci-lint v2.11.4) | `0 issues`, exit 0 |
| `make -j2 test` | exit 0; coverage file `_output/tests/linux_amd64/coverage.txt`; `config` and `internal/clients` both `ok` |
| `scripts/check-groups.sh` | 7 groups consistent across 6 sources |
| `scripts/check-workflows.sh` | pass; negative tests (expression in `run:`, unpinned action) fail as intended |
| `scripts/check-residue.sh` | no template residue |
| `actionlint` (with shellcheck) | exit 0, no findings |
| `make build.family` (xpkg dry-run) | exit 0, about 1.5 min: 8 binaries, 8 images (local Docker only), 8 xpkgs plus the monolith. Re-run at `c54bfbe` after the image fixes below: exit 0 |
| Gate re-run at `b239f0b` (generate, check_diff, build, lint, test, guards) | all exit 0; `0 issues`; check_diff clean; `actionlint` plain run exit 0 |
| Image build with a wrong OpenTofu SHA-256 | fails at `sha256sum -c` ("computed checksums did NOT match"), no image produced |

Notes on the toolchain: `make lint` and `make test` must run with Go 1.26.x (`GOTOOLCHAIN=go1.26.1`); with Go 1.27 the linter panics ("file requires newer Go version go1.27"). `goimports` must be on `PATH` for `make generate`; CI and `go.mod` now resolve the same `golang.org/x/tools` version (v0.49.0), so `go install` in CI reads it from `go.mod`.

Makefile fixes found by this verification (commit `eb225f1`):

- The schema recipe runs `tofu` with `TMPDIR=.` under `-chdir`. The provider plugin binds a Unix socket under `TMPDIR` and socket paths are limited to 108 bytes, so a long scratch `TMPDIR` made `make generate` fail with "Failed to read any lines from plugin's stdout". The Makefile now sets the short path itself, whatever the caller's `TMPDIR`.
- The schema is written to a temporary file and moved into place only on success. The previous recipe truncated `config/schema.json` on failure, which deleted 53 committed CRDs in one failed run. Regenerating after the fix is idempotent (`check_diff` clean).

## Unit and contract tests

Statement coverage of package `config` is 100% and of `internal/clients` 92% (the ProviderConfig resolution paths are covered with a fake client, see the re-run section below). The per-group config packages show 0% because their code is only exercised through `GetProvider()` in package `config`.

| Test | Guards |
|---|---|
| `TestBuildConfiguration`, `TestParseCredentials` | skip-empty, missing/blank token, `https://` + `/api/` base URL, trimming, no value in errors (canary) |
| `TestIncludeListHasAllResources`, `TestResourcesMatchTheSchema` | the include list equals the 24 resources, and equals the pinned schema |
| `TestGetIDFn`, `TestGetExternalNameFn`, `TestSlugResourcesSeedSlugFromExternalName` | external-name behavior and missing-parameter errors |
| `TestEveryResourceHasAStateContract`, `TestExternalNameRoundTrip` | for all 24 resources the synthetic state id equals the stored provider id and the external name round-trips |
| `TestGroupAndKindContract` | group and Kind of every resource, both trees |
| `TestReferences`, `TestReferenceGroupsAreAcyclic`, `TestReferenceFieldsExistInCRDs` | reference table, no import cycle between groups, `*Ref` / `*Selector` present or absent as designed |
| `TestDeprecatedAttributesAreOmitted`, `TestNoPlainSecretsInCRDs` | omitted attributes; sensitive inputs are `*SecretRef`; no plain secret fields |
| `TestCRDsPerGroup` | 48 resource CRDs and the 5 root CRDs, scope, version, non-empty schema, size < 700 KiB |
| `TestExamplesValidateAgainstCRDs` | every example validates against the CRD schema (unknown fields, types, top-level required parameters); one cluster example per resource and one namespaced per group |

Tests that guard the contract were checked to fail when broken (examples with a wrong field, workflow scripts with injected violations, a wrong checksum). The first versions of the external-name and credentials tests were written before the implementation but were not run red before it; the later additions (https-only `base_url`) were run red then green.

## Re-run after the review and security fixes

Run on the commit that carries this report, with Go 1.26.1 and goimports on `PATH`. Logs live in the run scratch directory; nothing was pushed, tagged or published.

| Step | Result |
|---|---|
| `make generate`, then `git status --short` | exit 0, 0 changed files (idempotent) |
| `make build` | exit 0 |
| `make lint` | `0 issues` (a first run flagged `gocyclo` on the new `validBaseURL` and two `noctx` findings in the new tests; both fixed) |
| `make test` | exit 0 |
| `go test -cover` | `config` 100.0%, `internal/clients` 92.2% (was 21.8%) |
| `scripts/check-groups.sh`, `check-workflows.sh`, `check-residue.sh` | exit 0 each |
| `actionlint .github/workflows/*.yml` | exit 0 |
| `make build.family` | exit 0, 8 packages |

New tests and guards:

- `internal/clients/setup_test.go`: `TerraformSetupBuilder` with a fake client for cluster-scoped, namespaced and cluster-wide ProviderConfigs; missing ProviderConfig, missing reference, unknown kind, missing secret or key, empty, malformed and non-object JSON, missing token, invalid `base_url`. Every error case asserts that the credential canary is not echoed.
- `internal/clients/upjet-sentry_test.go`: `base_url` with userinfo, host-less, dot segments, query, fragment and a valid port plus path prefix. They failed before the stricter validation was written.
- `config/singleton_test.go`: the two id-less singletons keep an id-less Terraform schema and their CRD shape, the upjet build in use gates the synthetic state id on `hasTFID` (id-less schemas get no id). The fork-only empty-id guard and `patches/apply.sh` were removed in the upstream upgrade.
- `scripts/check-workflows.sh`: a job that reads `UPBOUND_TOKEN` or `WILDBIT_CI_APP_KEY` outside `environment: production` fails (checked against the pre-change workflow), and the checksum pins in the Makefile and `publish.yml` must match.
- The host OpenTofu download is verified against the pinned SHA-256; a wrong value aborts the install.

## CRD facts (spike S2)

53 CRDs: 48 resource CRDs and 5 root-group CRDs. Largest: `Alert` 148 KiB, `IssueAlert` 96 KiB, `CronMonitor` 90 KiB. No CRD above 700 KiB.

## Package dry-run (T13)

`_output/xpkg/linux_amd64/`, each about 110 MB, each embedding a runtime image with `tofu`, the provider binary and the upstream license. The table is for the xpkgs built at `eb225f1`; the `c54bfbe` rebuild produced the same eight packages (`*-16.gc54bfbe.dirty.xpkg`, dirty only because the OpenTofu change was uncommitted at build time). Older `*.g035c545.dirty` and `*.g8209cd5.dirty` xpkgs in the same directory are stale local artifacts from before these fixes and are not used.

| Package | CRDs | Group | dependsOn |
|---|---|---|---|
| provider-family-sentry | 5 | root | none |
| provider-sentry-alert | 8 | alert | family |
| provider-sentry-dashboard | 2 | dashboard | family |
| provider-sentry-integration | 4 | integration | family |
| provider-sentry-monitor | 6 | monitor | family |
| provider-sentry-organization | 8 | organization | family |
| provider-sentry-project | 16 | project | family |
| provider-sentry-team | 4 | team | family |

53 CRDs in total, each package carries only its own group. Image fixes found while checking the runtime image as the non-root user (uid 65532):

- The license directory `/usr/share/licenses/terraform-provider-sentry` was created by `ADD --checksum` without a traversable mode, so the LICENSE file (0644) could not be reached by uid 65532 ("Permission denied"). Fixed in `c54bfbe` by creating the directory with `install -d -m 0755` first.
- The OpenTofu release zip was unpacked whole into `/usr/local/bin`, leaving the upstream `CHANGELOG.md`, `LICENSE` and `README.md` beside the binaries. Fixed in `f6463a0`: only `tofu` is extracted to `/usr/local/bin`, and the OpenTofu MPL-2.0 LICENSE is placed in `/usr/share/licenses/opentofu/LICENSE`.

Verified on the `c54bfbe` team image as uid 65532: both LICENSE files are readable (exit 0); `/usr/local/bin` holds `provider`, `terraform` (symlink) and `tofu`; `tofu version` reports v1.11.5. `up xpkg` and `crossplane xpkg push` were not run.

## Not covered

- Anything that needs a Sentry token: create/observe/update/delete cycles, token scopes, the import flows, the id-less resources at runtime, the effect of the per-configure health check, `jsontypes` diffs.
- Server-side CRD admission and CEL cost (no cluster in scope); CRDs were validated structurally and by example only.
- `make local-deploy` (kind) and the CI workflows themselves (never executed; actionlint and the guardrail scripts only).
- Multi-arch images (only linux/amd64 is published; the arm64 checksums are pinned but the arm64 image was not built).

## Upgrade run (upjet v2.5.1, OpenTofu 1.13.2)

Host Go 1.27.3 driven with `GOTOOLCHAIN=go1.26.8`. `make generate` (OpenTofu 1.13.2 schema dump), `make vendor vendor.check`, `make lint` (golangci-lint v2.11.4, 0 issues), `make -j2 test` and `scripts/check-{groups,workflows,residue}.sh` all pass. New tests: `TestGetImportIDFn`, `TestImportIDRoundTrip`. The only CRD delta is the additive `status.lastHandledReconcileAt` on all 48 CRDs. Live validation against a Sentry server was not run.
