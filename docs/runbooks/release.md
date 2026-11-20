# Release notes and specifics: provider-upjet-sentry

The release and rollout procedure is shared across the Upjet providers. See the org-gitops runbook
`docs/runbooks/upjet-provider-release-and-rollout.md` (https://github.com/wildbitca/org-gitops).

## Repo specifics

- Terraform provider and OpenTofu versions are pinned in two places: `Makefile` and `.github/workflows/publish.yml`.
- Sentry also pins the SHA-256 of the tofu and provider zips and LICENSE; recompute them from the official `SHA256SUMS`.
- Related: `docs/architecture/release-and-security.md`.

## Version history, 2026-11-20

- v1.0.1: bugfix/dependency release.
- v1.1.0: feature release with the Terraform provider bump to 0.16.1 (and OpenTofu 1.13.9).
