#!/usr/bin/env bash
set -euo pipefail

KEYRING_BACKEND="${KEYRING_BACKEND:-test}"
MERCHANT_KEY="${MERCHANT_KEY:-merchant}"
BYX_REST="${BYX_REST:-http://127.0.0.1:1317}"

ok() { echo "[OK] $*"; }
warn() { echo "[WARN] $*"; }
fail() { echo "[FAIL] $*"; exit 1; }
need() { command -v "$1" >/dev/null 2>&1 || fail "missing command: $1"; }

need byxd
need curl
need jq

key_address() {
  local name=$1
  byxd keys show "$name" -a --keyring-backend "$KEYRING_BACKEND" 2>/dev/null || true
}

create_key_if_missing() {
  local name=$1
  local existing
  existing="$(key_address "$name")"
  if [[ -n "$existing" ]]; then
    ok "key already exists: $name address=$existing"
    return 0
  fi

  warn "creating key '$name' in keyring backend '$KEYRING_BACKEND'"
  if ! byxd keys add "$name" --keyring-backend "$KEYRING_BACKEND" >/dev/null 2>&1; then
    fail "could not create key '$name'. Check keyring backend and local CLI setup."
  fi

  existing="$(key_address "$name")"
  [[ -n "$existing" ]] || fail "key '$name' was created but address could not be resolved"
  ok "key created: $name address=$existing"
}

print_balance_hint() {
  local addr=$1
  local amount
  amount="$(curl -sf "$BYX_REST/cosmos/bank/v1beta1/balances/$addr" 2>/dev/null | jq -r '(.balances[]? | select(.denom=="ubyx") | .amount) // "0"' || echo "0")"
  ok "merchant balance (ubyx): $amount"
}

create_key_if_missing "$MERCHANT_KEY"
merchant_addr="$(key_address "$MERCHANT_KEY")"
ok "merchant address: $merchant_addr"
print_balance_hint "$merchant_addr"
warn "fund this address in devnet (test funds only) before running E2E."
warn "no seed phrase or private key is printed by this script."
