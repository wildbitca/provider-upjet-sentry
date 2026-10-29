#!/usr/bin/env bash
# Verifies that the API group list is identical everywhere it is declared.
# The group names are a public contract; a drift between these lists means a
# group is generated but never published, or published without CRDs.
#
# Sources compared:
#   1. config/cluster/<group>/            (config packages, cluster tree)
#   2. config/namespaced/<group>/         (config packages, namespaced tree)
#   3. config/provider.go                 (<group>Cluster / <group>Namespaced imports)
#   4. cmd/provider/<group>/              (generated mains; skipped when absent)
#   5. publish.yml `REPOS=(...)`          (provider-sentry-<group>)
#   6. publish.yml `matrix.subpackage`
set -euo pipefail

cd "$(dirname "$0")/.."

PUBLISH=.github/workflows/publish.yml
fail=0

norm() { sort -u | grep -v '^$' || true; }

dirs() { find "$1" -mindepth 1 -maxdepth 1 -type d -printf '%f\n' | grep -vx common | norm; }

declare -A lists
lists[config/cluster]="$(dirs config/cluster)"
lists[config/namespaced]="$(dirs config/namespaced)"
lists[provider.go:cluster]="$(sed -nE 's#^\s*([a-z]+)Cluster "[^"]+/config/cluster/[a-z]+"$#\1#p' config/provider.go | norm)"
lists[provider.go:namespaced]="$(sed -nE 's#^\s*([a-z]+)Namespaced "[^"]+/config/namespaced/[a-z]+"$#\1#p' config/provider.go | norm)"
lists[publish.yml:REPOS]="$(sed -nE 's#^\s*provider-sentry-([a-z]+)\s*$#\1#p' "$PUBLISH" | norm)"
lists[publish.yml:matrix]="$(awk '/subpackage:/{f=1;next} f&&/^[[:space:]]*- /{print $2;next} f{f=0}' "$PUBLISH" | norm)"
if [ -d cmd/provider ]; then
  lists[cmd/provider]="$(dirs cmd/provider | grep -vxE 'config|monolith' | norm)"
fi

if ! grep -qE '^\s*provider-family-sentry\s*$' "$PUBLISH"; then
  echo "provider-family-sentry is missing from the REPOS array of $PUBLISH" >&2
  fail=1
fi

ref="${lists[config/cluster]}"
if [ -z "$ref" ]; then
  echo "no groups found under config/cluster" >&2
  exit 1
fi

for k in "${!lists[@]}"; do
  if [ "${lists[$k]}" != "$ref" ]; then
    echo "group list mismatch in $k:" >&2
    diff <(echo "$ref") <(echo "${lists[$k]}") >&2 || true
    fail=1
  fi
done

if [ "$fail" -eq 0 ]; then
  echo "group lists consistent ($(echo "$ref" | wc -l) groups: $(echo "$ref" | tr '\n' ' '))"
fi
exit "$fail"
