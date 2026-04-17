#!/usr/bin/env bash
# Phase 4 standalone smoke test.
#
# Creates one VM via rancher-machine, sshs into it, destroys it.
# REAL cloud resources are created — that costs real money. Read before running.
#
# Authentication: pick ONE of these two modes.
#
#   Mode A — Username/Password/Domain (recommended for Swiss OTC):
#     OS_USERNAME     IAM user name
#     OS_PASSWORD     IAM password
#     OS_DOMAIN_NAME  IAM domain (account root name)
#
#   Mode B — AK/SK (works on Standard OTC; on Swiss OTC the token catalog
#   lookup requires an explicit project, see OS_PROJECT_NAME below):
#     OS_ACCESS_KEY
#     OS_SECRET_KEY
#
# Required for both modes:
#   OS_REGION       eu-ch2 | eu-de | eu-nl
#
# Required for Swiss OTC (and recommended for multi-project Standard OTC):
#   OS_PROJECT_NAME project to scope the token to (e.g. eu-ch2_wotest)
#
# Optional:
#   MACHINE_NAME        defaults to "smoke-<region>"
#   OS_IMAGE_NAME       override the driver's default Ubuntu image
#   SKIP_DESTROY        if set (any value), leaves the VM up for inspection

set -euo pipefail

die() {
  printf 'FAIL: %s\n' "$*" >&2
  exit 2
}

# Determine which auth mode we have.
have_userpass=0
have_aksk=0
if [[ -n "${OS_USERNAME:-}" && -n "${OS_PASSWORD:-}" && -n "${OS_DOMAIN_NAME:-}" ]]; then
  have_userpass=1
fi
if [[ -n "${OS_ACCESS_KEY:-}" && -n "${OS_SECRET_KEY:-}" ]]; then
  have_aksk=1
fi

if (( have_userpass == 0 && have_aksk == 0 )); then
  die "no auth method configured — set OS_USERNAME/OS_PASSWORD/OS_DOMAIN_NAME or OS_ACCESS_KEY/OS_SECRET_KEY"
fi

[[ -n "${OS_REGION:-}" ]] || die "OS_REGION is required (eu-ch2 / eu-de / eu-nl)"

if [[ "$OS_REGION" == "eu-ch2" && -z "${OS_PROJECT_NAME:-}" ]]; then
  die "Swiss OTC (eu-ch2) requires OS_PROJECT_NAME — unscoped tokens have no service catalog"
fi

MACHINE_NAME="${MACHINE_NAME:-smoke-${OS_REGION}}"
REPO_ROOT="$(git -C "$(dirname "$0")" rev-parse --show-toplevel)"

echo "=== [1/5] Building driver binary ==="
(cd "$REPO_ROOT" && make build)

BIN="$REPO_ROOT/bin/docker-machine-driver-opentelekomcloud"
[[ -x "$BIN" ]] || die "expected binary at $BIN"

echo "=== [2/5] Ensuring binary is on PATH ==="
# rancher-machine resolves driver-name -> docker-machine-driver-<name> on PATH.
# We prepend ./bin so no sudo install is needed to try things out.
export PATH="$REPO_ROOT/bin:$PATH"

# Assemble rancher-machine flags based on detected auth mode.
flags=(
  --driver opentelekomcloud
  --opentelekomcloud-region "$OS_REGION"
)
if (( have_userpass )); then
  flags+=(
    --opentelekomcloud-username    "$OS_USERNAME"
    --opentelekomcloud-password    "$OS_PASSWORD"
    --opentelekomcloud-domain-name "$OS_DOMAIN_NAME"
  )
elif (( have_aksk )); then
  flags+=(
    --opentelekomcloud-access-key "$OS_ACCESS_KEY"
    --opentelekomcloud-secret-key "$OS_SECRET_KEY"
  )
fi
if [[ -n "${OS_PROJECT_NAME:-}" ]]; then
  flags+=( --opentelekomcloud-project-name "$OS_PROJECT_NAME" )
fi
if [[ -n "${OS_IMAGE_NAME:-}" ]]; then
  flags+=( --opentelekomcloud-image-name "$OS_IMAGE_NAME" )
fi

echo "=== [3/5] Creating VM '$MACHINE_NAME' in region '$OS_REGION' (project: ${OS_PROJECT_NAME:-<none>}) ==="
rancher-machine create "${flags[@]}" "$MACHINE_NAME"

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
