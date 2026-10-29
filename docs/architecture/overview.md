# Architecture overview

`provider-upjet-sentry` is an Upjet provider that wraps the Terraform provider `jianyuan/sentry` v0.15.8. It runs the unmodified provider binary through OpenTofu (the "CLI" path), like the sibling `provider-upjet-*` repositories (bunnynet, cloudflare and the template provider).

## Why the CLI path

- `jianyuan/sentry` is itself a mux provider (SDKv2 for 7 resources, Plugin Framework for 17). The release binary already contains the mux and speaks protocol 6, so OpenTofu needs no extra wiring.
- The no-fork path (`TerraformPluginFrameworkIncludeList`) would require importing the provider's Go packages. The Framework resources live under `internal/`, which the Go compiler forbids importing from another module, so 17 of 24 resources could never use it.
- Cost: one OpenTofu workspace per operation (more memory and latency than in-process), the same as our other providers.

Only `ujconfig.WithIncludeList` is used. Nothing in the repository imports `terraform-plugin-framework`, `terraform-plugin-sdk` (except `config/common`, to edit schemas) or `terraform-plugin-mux`.

## Data flow

1. `make generate` runs `tofu init -upgrade` and `tofu providers schema -json` for `jianyuan/sentry` to produce `config/schema.json`.
2. `cmd/generator` loads the schema, applies `config/` (groups, Kinds, references, external names, omitted attributes) and writes `apis/`, `internal/controller/` and `cmd/provider/<group>/`; controller-gen produces `package/crds/`.
3. At runtime a controller builds a Terraform workspace per managed resource, writes `main.tf.json` plus a synthetic state, and runs `tofu` against the native provider binary (`TERRAFORM_NATIVE_PROVIDER_PATH`).
4. Credentials come from the `ProviderConfig` Secret (see `credentials.md`).

## Repository layout

| Path | Content |
|---|---|
| `config/` | Hand-written configuration. `provider.go`, `external_name.go`, `common/` (shared references and `Omit`), `{cluster,namespaced}/<group>/config.go` |
| `apis/`, `internal/controller/`, `cmd/provider/` | Generated (`zz_*`), plus hand-written `ProviderConfig` types and controllers |
| `internal/clients/upjet-sentry.go` | `TerraformSetupBuilder`, credential parsing |
| `package/crds/` | Generated CRDs: 24 resources x 2 scopes + ProviderConfig kinds |
| `cluster/images/provider-upjet-sentry/` | Dockerfile (OpenTofu + provider binary, checksum verified) |
| `scripts/` | `build-family.sh`, `check-groups.sh`, `check-workflows.sh` |
| `examples/` | One manifest per resource |
| `docs/` | This documentation |

## Two API trees

Each group exists as a cluster-scoped API (`<group>.upjet-sentry.upbound.io`) and a namespaced API (`<group>.upjet-sentry.m.upbound.io`); the namespaced tree also has `ClusterProviderConfig`. Both trees share the same `config/common` helpers and mirror each other's `config.go`.

## Packaging

One family package (`provider-family-sentry`, ProviderConfig CRDs) plus one sub-provider per group (`provider-sentry-<group>`, seven in total), published to `xpkg.upbound.io/wildbitca/`, linux/amd64 only. Seven groups keep publish time low (publish cost scales with the group count). See `release-and-security.md`.
