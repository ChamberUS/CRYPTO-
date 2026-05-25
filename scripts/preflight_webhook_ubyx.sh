#!/usr/bin/env bash
set -euo pipefail

BYX_REST="${BYX_REST:-http://127.0.0.1:1317}"
BYX_RPC="${BYX_RPC:-http://127.0.0.1:26657}"
BYX_CHAIN_MODE="${BYX_CHAIN_MODE:-}"
KEYRING_BACKEND="${KEYRING_BACKEND:-test}"
MERCHANT_KEY="${MERCHANT_KEY:-merchant}"
PAYER_KEY="${PAYER_KEY:-payer}"
AMOUNT_UBYX="${AMOUNT_UBYX:-500000}"
MIN_MERCHANT_BALANCE_UBYX="${MIN_MERCHANT_BALANCE_UBYX:-1}"
MIN_PAYER_BALANCE_UBYX="${MIN_PAYER_BALANCE_UBYX:-$AMOUNT_UBYX}"
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
  echo "KEYRING_BACKEND=$KEYRING_BACKEND"
  echo "REQUIRED_KEYS=$MERCHANT_KEY,$PAYER_KEY"
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

normalize_uint() {
  local value=$1
  if [[ ! "$value" =~ ^[0-9]+$ ]]; then
    echo "0"
    return
  fi
  value="$(echo "$value" | sed -E 's/^0+//')"
  if [[ -z "$value" ]]; then
    echo "0"
  else
    echo "$value"
  fi
}

uint_gte() {
  local left right
  left="$(normalize_uint "$1")"
  right="$(normalize_uint "$2")"
  if (( ${#left} > ${#right} )); then
    return 0
  fi
  if (( ${#left} < ${#right} )); then
    return 1
  fi
  [[ "$left" > "$right" || "$left" == "$right" ]]
}

balance_ubyx_for_address() {
  local address=$1
  curl -sf "$BYX_REST/cosmos/bank/v1beta1/balances/$address" \
    | jq -r '(.balances[]? | select(.denom=="ubyx") | .amount) // "0"'
}

check_key_and_balance() {
  local key_name=$1
  local min_balance=$2
  local address balance

  address="$(byxd keys show "$key_name" -a --keyring-backend "$KEYRING_BACKEND" 2>/dev/null || true)"
  if [[ -z "$address" ]]; then
    fail "key '$key_name' not found in keyring backend '$KEYRING_BACKEND'. Run: make e2e-webhook-ubyx-keys"
  fi

  ok "key available: $key_name address=$address"
  balance="$(balance_ubyx_for_address "$address")"
  balance="$(normalize_uint "$balance")"
  ok "key balance: $key_name ubyx=$balance"

  if ! uint_gte "$balance" "$min_balance"; then
    fail "insufficient ubyx for key '$key_name' (address $address): have $balance, need at least $min_balance. Fund in devnet and retry."
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

if [[ "$MERCHANT_KEY" == "$PAYER_KEY" ]]; then
  required_balance="$MIN_PAYER_BALANCE_UBYX"
  if ! uint_gte "$required_balance" "$MIN_MERCHANT_BALANCE_UBYX"; then
    required_balance="$MIN_MERCHANT_BALANCE_UBYX"
  fi
  check_key_and_balance "$MERCHANT_KEY" "$required_balance"
else
  check_key_and_balance "$MERCHANT_KEY" "$MIN_MERCHANT_BALANCE_UBYX"
  check_key_and_balance "$PAYER_KEY" "$MIN_PAYER_BALANCE_UBYX"
fi

echo "Preflight webhook/ubyx passed."
