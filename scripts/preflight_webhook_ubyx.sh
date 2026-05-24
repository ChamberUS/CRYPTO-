#!/usr/bin/env bash
set -euo pipefail

BYX_REST="${BYX_REST:-http://127.0.0.1:1317}"
BYX_RPC="${BYX_RPC:-http://127.0.0.1:26657}"
BYX_CHAIN_MODE="${BYX_CHAIN_MODE:-}"
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
E2E_DIR="${E2E_DIR:-$ROOT_DIR/.e2e/webhook-ubyx}"

ok() { echo "[OK] $*"; }
warn() { echo "[WARN] $*"; }
fail() { echo "[FAIL] $*"; exit 1; }
need() { command -v "$1" >/dev/null 2>&1 || fail "missing command: $1"; ok "command available: $1"; }

normalize_chain_mode() {
  if [[ -n "$BYX_CHAIN_MODE" ]]; then
    case "$BYX_CHAIN_MODE" in
      external|byxd|custom|ignite) return 0 ;;
      *) fail "invalid BYX_CHAIN_MODE='$BYX_CHAIN_MODE' (use: external|byxd|custom|ignite)" ;;
    esac
  fi

  if curl -sf "$BYX_REST/cosmos/base/tendermint/v1beta1/syncing" >/dev/null 2>&1 && curl -sf "$BYX_RPC/status" >/dev/null 2>&1; then
    BYX_CHAIN_MODE="external"
    warn "BYX_CHAIN_MODE not set; inferred external because REST/RPC are already healthy"
    return 0
  fi

  warn "BYX_CHAIN_MODE not set and REST/RPC not healthy."
  warn "Set one of: BYX_CHAIN_MODE=external|byxd|custom|ignite"
  BYX_CHAIN_MODE="(unset)"
}

print_mode_guidance() {
  local attempted_start="no"
  case "$BYX_CHAIN_MODE" in
    external)
      attempted_start="no"
      ;;
    byxd|custom|ignite)
      attempted_start="yes (expected via stack-webhook-ubyx-up)"
      ;;
    *)
      attempted_start="unknown (BYX_CHAIN_MODE unset)"
      ;;
  esac

  echo "BYX_CHAIN_MODE=$BYX_CHAIN_MODE"
  echo "BYX_REST=$BYX_REST"
  echo "BYX_RPC=$BYX_RPC"
  echo "CHAIN_START_ATTEMPT_EXPECTED=$attempted_start"
  echo "PORT_DIAG_REST_1317=endpoint:$BYX_REST"
  echo "PORT_DIAG_RPC_26657=endpoint:$BYX_RPC"
}

print_ignite_buf_hint_if_needed() {
  local chain_log="$E2E_DIR/chain.log"
  if [[ "$BYX_CHAIN_MODE" == "ignite" ]] && [[ -f "$chain_log" ]] && grep -q "buf\\.build" "$chain_log"; then
    echo "Ignite mode failed while trying to access buf.build."
    echo "This is an environment/network/proto-cache issue."
    echo "Use BYX_CHAIN_MODE=external for an already running chain,"
    echo "BYX_CHAIN_MODE=byxd for a built binary,"
    echo "or BYX_CHAIN_MODE=custom with BYX_CHAIN_START_CMD."
  fi
}

need byxd
need curl
need jq
need openssl
need node
need npm

normalize_chain_mode
print_mode_guidance

if curl -sf "$BYX_REST/cosmos/base/tendermint/v1beta1/syncing" >/dev/null; then
  ok "REST reachable: $BYX_REST"
else
  print_ignite_buf_hint_if_needed
  fail "REST unavailable: $BYX_REST (expected REST on port 1317)"
fi

if curl -sf "$BYX_RPC/status" >/dev/null; then
  ok "RPC reachable: $BYX_RPC"
else
  print_ignite_buf_hint_if_needed
  fail "RPC unavailable: $BYX_RPC (expected RPC on port 26657)"
fi

# byxd status should also work in local CLI context
byxd status >/dev/null 2>&1 \
  && ok "byxd status available" \
  || fail "byxd status failed"
echo "Preflight webhook/ubyx passed."
