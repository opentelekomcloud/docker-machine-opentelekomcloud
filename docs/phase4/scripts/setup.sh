#!/usr/bin/env bash
# Phase 4 one-shot environment setup.
#
# Ensures the host has the two moving pieces needed to run smoke-test.sh:
#
#   1. rancher-machine — Rancher-flavoured Docker Machine CLI. Rancher only
#      ships Linux binaries; on macOS we build from source under
#      ~/.local/bin.
#
#   2. docker-machine-driver-opentelekomcloud — our driver binary, built from
#      this repo into ./bin/ via `make build`.
#
# Safe to run repeatedly; each step short-circuits when the artefact is
# already present and current.

set -euo pipefail

RANCHER_MACHINE_VERSION="v0.15.0-rancher142"
RANCHER_MACHINE_BIN="${HOME}/.local/bin/rancher-machine"
REPO_ROOT="$(git -C "$(dirname "$0")" rev-parse --show-toplevel)"

log() { printf '\033[1m[setup]\033[0m %s\n' "$*"; }

# --- rancher-machine ---
if command -v rancher-machine >/dev/null 2>&1; then
  log "rancher-machine already on PATH at $(command -v rancher-machine)"
elif [[ -x "$RANCHER_MACHINE_BIN" ]]; then
  log "rancher-machine present at $RANCHER_MACHINE_BIN (add \$HOME/.local/bin to PATH)"
else
  case "$(uname -s)" in
    Darwin)
      log "building rancher-machine $RANCHER_MACHINE_VERSION from source (Rancher ships Linux-only binaries)"
      command -v go >/dev/null || { echo "Go toolchain required to build rancher-machine on macOS"; exit 2; }
      mkdir -p "$(dirname "$RANCHER_MACHINE_BIN")"
      src="$(mktemp -d)/rancher-machine"
      git clone --depth 1 --branch "$RANCHER_MACHINE_VERSION" \
        https://github.com/rancher/machine.git "$src" >/dev/null
      (cd "$src" && go build -o "$RANCHER_MACHINE_BIN" ./cmd/rancher-machine)
      rm -rf "$(dirname "$src")"
      log "rancher-machine installed at $RANCHER_MACHINE_BIN"
      ;;
    Linux)
      log "downloading rancher-machine $RANCHER_MACHINE_VERSION (Linux binary)"
      mkdir -p "$(dirname "$RANCHER_MACHINE_BIN")"
      arch="$(uname -m)"
      case "$arch" in
        aarch64|arm64) url_arch="arm64" ;;
        x86_64|amd64)  url_arch="amd64" ;;
        *) echo "unsupported arch: $arch"; exit 2 ;;
      esac
      curl -fsSL -o /tmp/rm.tgz \
        "https://github.com/rancher/machine/releases/download/${RANCHER_MACHINE_VERSION}/rancher-machine-${url_arch}.tar.gz"
      tar -xzf /tmp/rm.tgz -C /tmp
      mv /tmp/rancher-machine "$RANCHER_MACHINE_BIN"
      chmod +x "$RANCHER_MACHINE_BIN"
      rm /tmp/rm.tgz
      log "rancher-machine installed at $RANCHER_MACHINE_BIN"
      ;;
    *) echo "unsupported OS: $(uname -s)"; exit 2 ;;
  esac
fi

# --- driver binary ---
log "building driver binary"
(cd "$REPO_ROOT" && make build)
if [[ ! -x "$REPO_ROOT/bin/docker-machine-driver-opentelekomcloud" ]]; then
  echo "driver build produced no binary at $REPO_ROOT/bin/" >&2
  exit 3
fi
log "driver binary at $REPO_ROOT/bin/docker-machine-driver-opentelekomcloud"

# --- env hint ---
log "done. Next: export PATH=\"\$HOME/.local/bin:\$REPO_ROOT/bin:\$PATH\" then run smoke-test.sh"
log "or one-shot with 1Password creds:  op run --env-file=$REPO_ROOT/docs/phase4/scripts/.env.op -- $REPO_ROOT/docs/phase4/scripts/smoke-test.sh"
