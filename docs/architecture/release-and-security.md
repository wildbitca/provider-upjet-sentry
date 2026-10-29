# Release and security

## Release flow

1. `Tag` workflow (`workflow_dispatch` from `main` only, input `version` such as `v1.0.0`, validated as semver; runs under environment `production` and re-checks through the API that the dispatched commit is on `main` before it reads the App key) pushes an annotated tag using a short-lived GitHub App token (`vars.WILDBIT_CI_APP_ID`, `secrets.WILDBIT_CI_APP_KEY`); the default token is blocked by the tag ruleset.
2. `Publish` runs on tags `v*`: `prepare` (checks out `refs/tags/<version>`, requires the tag to be reachable from `main`; `workflow_dispatch` is only accepted from `main` or a release tag) -> `ensure-repos` -> `family-config` -> `family-subproviders` (matrix, seven groups, `max-parallel: 5`) -> `release`. Every job that reads `UPBOUND_TOKEN` runs under environment `production`, and every build job checks out the same tag ref. `scripts/check-workflows.sh` fails when a job reads a publishing secret outside that environment.
3. `ensure-repos` creates the Upbound repositories only when they do not exist (the repository update API is rate limited) and fails fast when `UPBOUND_TOKEN` is missing.
4. Platform: linux/amd64 only. Registry: `xpkg.upbound.io/wildbitca/`.

Publishing, tagging and setting secrets are **not** part of the implementation scope and require explicit approval.

## Group sync rule

The group list is declared in six places and must stay identical: `config/cluster/<group>/`, `config/namespaced/<group>/`, the imports in `config/provider.go`, `cmd/provider/*` (generated), and `publish.yml` (`REPOS` array and `matrix.subpackage`); the `REPOS` array must also list `provider-family-sentry`. `scripts/check-groups.sh` verifies it and runs in CI (`guardrails` job, together with `check-workflows.sh` and `check-residue.sh`, which rejects template text such as the old credential keys in shipped files). A new group means: add it everywhere in one change, and never move an existing resource between groups after a release.

## Pinning policy

- GitHub Actions: full commit SHAs with a version comment. `scripts/check-workflows.sh` fails on any other form.
- Tools downloaded in workflows: version plus SHA-256 (`up` v0.54.0, `docker-credential-up`, Crossplane CLI v2.1.3 checksum `fe3ac3ab...0013`, the published `crank.sha256`). No `curl | sh` anywhere (the template workflow's unpinned installer was deliberately not copied).
- Images: `alpine:3.23.3` pinned by digest. The OpenTofu and provider archives and the upstream license text are verified against SHA-256 values committed in this repository (`Makefile`, mirrored in `publish.yml`; verified in the Dockerfile by `sha256sum -c` and `ADD --checksum`). An empty or wrong value fails the image build. The values come from the `SHA256SUMS` published next to each release; **when the provider or OpenTofu version changes, update the hashes in both files in the same change**. The signatures of the upstream release checksum files are not verified.
- Dependabot: weekly, 7-day cooldown, for Go modules, GitHub Actions and the Dockerfile base image.
- `golang.org/x/net` >= v0.60.0 (five HTTP/2 CVEs fixed in earlier bumps of the sibling repos).

## Workflow hardening

- Default `permissions: contents: read`; `contents: write` only in the `release` job.
- Secrets and every other expression are passed through `env:`, never interpolated into `run:` scripts (`check-workflows.sh` enforces it). The release version is validated as semver in `tag.yml` and `publish.yml`.
- The publish jobs do not use `actions/cache` or the GitHub Actions build cache, so a poisoned cache entry cannot reach a release build.
- `ci.yml` has read-only token permissions; only the `lint` job gets `pull-requests: write` and `checks: write` for annotations.
- **Secret scope.** `UPBOUND_TOKEN` and `WILDBIT_CI_APP_KEY` are read only by jobs that run under environment `production`, so the environment's deployment protection rules and branch/tag restrictions apply to every use. Consequence: a protection rule that needs a manual approval fires once per job (and per matrix entry); keep the rule for this environment non-interactive or approve once per run. Restrict the environment to protected branches and tags in the repository settings.
- CodeQL on push, PR and weekly. Branch protection and the tag ruleset exceptions belong to the CI App only.

## Public repository hygiene

- Examples use placeholders only. No token or DSN appears in the tree.
- `SECURITY.md` points to GitHub private vulnerability reporting.
- The `secret-scan` CI job runs a pinned gitleaks binary (checksum verified) over the full history. Run it locally over the history before the first push.

## Licensing

Apache-2.0 for this repository. The wrapped provider is MIT; its license ships in the image at `/usr/share/licenses/terraform-provider-sentry/LICENSE`, and `NOTICE` records the attribution.

## Attribution rule

No `Co-Authored-By` or assistant attribution in commits or PR descriptions; commit messages are in English.

## Known follow-ups (not blocking)

- The image does not carry this repository's `NOTICE` or the OpenTofu license text/source offer; only the upstream provider license is installed.
- Signatures of the upstream release checksum files are not verified (the pinned hashes are).
- `.github/CODEOWNERS` assigns `@wildbitca/admin`; enable "Require review from Code Owners" in the branch ruleset.
- How often the `production` environment protection fires per run (once per job) is unverified; see the secret scope note above.
- Branch protection, tag ruleset, CodeQL, private vulnerability reporting and Dependabot alerts are GitHub settings and are configured at repository creation (T17).

## Package icon

`package/icon.svg`, `extensions/icons/icon.svg` and the `iconURI` data URI in `package/crossplane.yaml` use the Sentry glyph from `static/images/logo-sentry.svg` of https://github.com/getsentry/sentry (the asset behind https://sentry.io/branding/), white on the brand purple `#362d59`, viewBox 80x80. The Sentry source is licensed FSL-1.1-Apache-2.0, but the name and logo are trademarks of Functional Software, Inc. dba Sentry; this provider is not affiliated with or endorsed by Sentry, and the logo is used only to identify the product the provider manages. The marketplace icon comes from `up xpkg append --extensions-root=./extensions`; the inline `crossplane.yaml` heredocs in `publish.yml` do not carry `iconURI` (same as the sibling providers).
