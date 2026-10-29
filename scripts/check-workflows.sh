#!/usr/bin/env bash
# Supply-chain guardrails for .github/: no piped installers, every third-party
# action is pinned to a full commit SHA, and no GitHub expression is interpolated
# into a `run:` script (pass values through `env:` instead).
set -euo pipefail

cd "$(dirname "$0")/.."

fail=0

if grep -rnE '\|\s*(sudo\s+)?(ba)?sh\b' .github/; then
  echo "found a piped shell installer in .github/ (curl | sh); download a pinned release and verify its checksum" >&2
  fail=1
fi

if grep -rnE '^\s*-?\s*uses:\s*[^./][^@]*@' .github/workflows | grep -vE '@[0-9a-f]{40}(\s|$)'; then
  echo "found an action that is not pinned to a full commit SHA" >&2
  fail=1
fi

# `${{ ... }}` inside a run block is script injection waiting to happen.
for f in .github/workflows/*.y*ml; do
  if awk -v file="$f" '
    function indent(l) { match(l, /^ */); return RLENGTH }
    {
      if (inrun) {
        if ($0 ~ /^[[:space:]]*$/) next
        if (indent($0) <= runindent) { inrun = 0 }
        else if ($0 ~ /\$\{\{/) { printf "%s:%d: expression inside run: %s\n", file, NR, $0; bad = 1 }
      }
      if (!inrun && $0 ~ /^[[:space:]]*(- )?run:/) {
        runindent = indent($0)
        if ($0 ~ /\$\{\{/) { printf "%s:%d: expression inside run: %s\n", file, NR, $0; bad = 1 }
        if ($0 ~ /run:[[:space:]]*[|>]/ || $0 ~ /run:[[:space:]]*$/) inrun = 1
      }
    }
    END { exit bad }
  ' "$f"; then :; else
    echo "pass GitHub expressions to run: scripts through env:, not by interpolation" >&2
    fail=1
  fi
done

# The image checksums are pinned in the Makefile and again in publish.yml's env: block;
# a bump must touch both.
for var in OPENTOFU_SHA256_AMD64 OPENTOFU_SHA256_ARM64 TERRAFORM_PROVIDER_SHA256_AMD64 \
  TERRAFORM_PROVIDER_SHA256_ARM64 TERRAFORM_PROVIDER_LICENSE_SHA256; do
  mk=$(sed -nE "s/^export ${var} \?= *([0-9a-f]{64})\s*$/\1/p" Makefile)
  wf=$(sed -nE "s/^  ${var}: *'([0-9a-f]{64})'\s*$/\1/p" .github/workflows/publish.yml)
  if [ -z "${mk}" ] || [ "${mk}" != "${wf}" ]; then
    echo "${var} differs between Makefile and publish.yml (or is missing)" >&2
    fail=1
  fi
done

# Jobs that read a publishing secret must run under the production environment
# (deployment protection rules and branch restrictions apply there).
for f in .github/workflows/*.y*ml; do
  if awk -v file="$f" '
    function flush() {
      if (job != "" && uses && !prod) { printf "%s: job %s reads a publishing secret outside environment: production\n", file, job; bad = 1 }
    }
    /^[^ #]/ { injobs = ($0 ~ /^jobs:/) }
    injobs && /^  [A-Za-z0-9_-]+:[[:space:]]*$/ { flush(); job = $1; uses = 0; prod = 0; next }
    injobs && /secrets\.(UPBOUND_TOKEN|WILDBIT_CI_APP_KEY)/ { uses = 1 }
    injobs && /^    environment:[[:space:]]*production[[:space:]]*$/ { prod = 1 }
    END { flush(); exit bad }
  ' "$f"; then :; else
    fail=1
  fi
done

exit "$fail"
