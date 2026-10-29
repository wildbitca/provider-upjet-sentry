# provider-upjet-sentry (Crossplane / Upjet provider)

## Documentation

Write durable documentation directly into the committed docs tree, never into a gitignored scratch path:

- SDDs / designs: `docs/sdd/`
- ADRs / architecture decisions and reference docs: `docs/architecture/`
- Open items and status: `docs/architecture/status.md`

## Conventions

- Everything written in this repository is in English: code, comments, commits, docs.
- Commits carry no assistant attribution (no `Co-Authored-By`, no "generated with").
- Never add `/config` to `.gitignore`: it hides new `config/<scope>/<group>` packages (cloudflare `b253056`).
- API group and Kind names are a public contract. Never rename them or move a resource between groups after a release. `config/provider_test.go` enforces the table.
- The group list must be identical in `config/{cluster,namespaced}/`, `config/provider.go`, `cmd/provider/*` (generated), and `.github/workflows/publish.yml` (`REPOS` array and `matrix.subpackage`). `scripts/check-groups.sh` checks it.
- Every `ExternalNameConfigs` entry is also the include list: a resource missing there silently does not compile.
- Keep `go 1.26.8` (the floor upjet v2.5 sets) in `go.mod`; check the `go` directive before committing after any `go mod tidy`. Golangci-lint v2.11.4 panics on a Go 1.27 host: run lint and generation with `GOTOOLCHAIN=go1.26.8` and `goimports` on `PATH`.
- Upjet is pinned to the `wildbitca/upjet` fork tag via `replace` (see `docs/architecture/upjet-fork.md`). Observe-only imports use `ExternalName.GetImportIDFn` (`withImportID` in `config/external_name.go`), not `GetIDFn`; see `docs/architecture/external-names.md`.
- Do not hand-edit generated files (`zz_*`, `package/crds/`, `config/schema.json`). Run `make generate`; before regenerating, delete stale `zz_*` files if a config change removed a cross-group reference (a stale resolver can create an import cycle that blocks controller-gen).
- Pin everything: GitHub Actions to full SHAs, downloaded tools to a version plus SHA-256. `scripts/check-workflows.sh` rejects `curl | sh`, unpinned actions and expressions inside `run:` blocks; `scripts/check-residue.sh` rejects template text (names and credential keys of the provider this repo was bootstrapped from) in shipped files.

## Building

See `README.md` (Development). Tooling: Go 1.26+, OpenTofu 1.13.2 (fetched by the Makefile), `goimports` on `PATH`, Docker for images.
- Image checksums: the SHA-256 values for OpenTofu, the provider zips and the upstream LICENSE live in `Makefile` and `.github/workflows/publish.yml`. Bump them together with `OPENTOFU_VERSION` / `TERRAFORM_PROVIDER_VERSION`.
