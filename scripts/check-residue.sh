#!/usr/bin/env bash
# Fails when text from the provider template (wording of the bootstrap template, its credential
# keys) is still present in hand-written files that ship with the packages.
set -euo pipefail

cd "$(dirname "$0")/.."

paths=(extensions .github cluster hack internal/clients config/common config/cluster config/namespaced \
  config/provider.go config/external_name.go Makefile upbound.yaml package/crossplane.yaml README.md \
  CHANGELOG.md docs/family examples)

if grep -rniE 'supa''base|access_token|YOUR_.*_TOKEN' "${paths[@]}" 2>/dev/null; then
  echo "template residue found (see matches above)" >&2
  exit 1
fi
echo "no template residue"
