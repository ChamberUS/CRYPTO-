#!/usr/bin/env bash
set -euo pipefail

BYX_REST="${BYX_REST:-http://127.0.0.1:1317}"
BYX_RPC="${BYX_RPC:-http://127.0.0.1:26657}"

ok() { echo "[OK] $*"; }
warn() { echo "[WARN] $*"; }
fail() { echo "[FAIL] $*"; exit 1; }
need() { command -v "$1" >/dev/null 2>&1 || fail "missing command: $1"; ok "command available: $1"; }

need byxd
need curl
need jq
need openssl
need node
need npm

curl -sf "$BYX_REST/cosmos/base/tendermint/v1beta1/syncing" >/dev/null \
  && ok "REST reachable: $BYX_REST" \
  || fail "REST unavailable: $BYX_REST"

curl -sf "$BYX_RPC/status" >/dev/null \
  && ok "RPC reachable: $BYX_RPC" \
  || fail "RPC unavailable: $BYX_RPC"

# byxd status should also work in local CLI context
byxd status >/dev/null 2>&1 \
  && ok "byxd status available" \
  || fail "byxd status failed"

echo "Preflight webhook/ubyx passed."
