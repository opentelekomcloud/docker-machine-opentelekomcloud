#!/usr/bin/env bash
# Phase 4 standalone smoke test.
#
# Creates one VM via rancher-machine, sshs into it, destroys it.
# REAL cloud resources are created — that costs real money. Read before running.
#
# Required env:
#   OS_ACCESS_KEY   Swiss OTC (or Standard) access key
#   OS_SECRET_KEY   matching secret key
#   OS_REGION       eu-ch2 | eu-de | eu-nl
#
# Optional env:
#   MACHINE_NAME    defaults to "smoke-<region>"
#   SKIP_DESTROY    if set (any value), leaves the VM up for inspection

set -euo pipefail

require() {
  local var="$1"
  if [[ -z "${!var:-}" ]]; then
    printf 'FAIL: env var %s is required but not set\n' "$var" >&2
    exit 2
  fi
}

require OS_ACCESS_KEY
require OS_SECRET_KEY
require OS_REGION

MACHINE_NAME="${MACHINE_NAME:-smoke-${OS_REGION}}"
REPO_ROOT="$(git -C "$(dirname "$0")" rev-parse --show-toplevel)"

echo "=== [1/5] Building driver binary ==="
(cd "$REPO_ROOT" && make build)

BIN="$REPO_ROOT/bin/docker-machine-driver-opentelekomcloud"
if [[ ! -x "$BIN" ]]; then
  echo "FAIL: expected binary at $BIN" >&2
  exit 3
fi

echo "=== [2/5] Ensuring binary is on PATH ==="
# rancher-machine resolves driver-name -> docker-machine-driver-<name> on PATH.
# We prepend ./bin so no sudo install is needed to try things out.
export PATH="$REPO_ROOT/bin:$PATH"

echo "=== [3/5] Creating VM '$MACHINE_NAME' in region '$OS_REGION' ==="
rancher-machine create \
  --driver opentelekomcloud \
  --opentelekomcloud-region "$OS_REGION" \
  --opentelekomcloud-access-key "$OS_ACCESS_KEY" \
  --opentelekomcloud-secret-key "$OS_SECRET_KEY" \
  "$MACHINE_NAME"

echo "=== [4/5] SSH probe ==="
rancher-machine ssh "$MACHINE_NAME" "uname -a && cat /etc/os-release | head -5"

if [[ -n "${SKIP_DESTROY:-}" ]]; then
  echo "=== [5/5] SKIP_DESTROY set — leaving '$MACHINE_NAME' running. Remember to clean up. ==="
  echo
  echo "PASS (partial — no destroy verification)"
  exit 0
fi

echo "=== [5/5] Destroying VM '$MACHINE_NAME' ==="
rancher-machine rm -f "$MACHINE_NAME"

echo
echo "PASS: create + ssh + destroy cycle completed for region $OS_REGION"
echo
echo "Manual follow-up: audit OTC for any orphan resources (VPC, SG, EIP) tied to $MACHINE_NAME"
