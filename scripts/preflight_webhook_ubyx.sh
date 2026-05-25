#!/usr/bin/env bash
set -euo pipefail

BYX_REST="${BYX_REST:-http://127.0.0.1:1317}"
BYX_RPC="${BYX_RPC:-http://127.0.0.1:26657}"
BYX_CHAIN_MODE="${BYX_CHAIN_MODE:-}"
BYXD_BIN="${BYXD_BIN:-byxd}"
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

resolve_byxd_bin() {
  if [[ "$BYXD_BIN" == */* ]]; then
    [[ -x "$BYXD_BIN" ]] || fail "BYXD_BIN is not executable: $BYXD_BIN"
    echo "$BYXD_BIN"
    return 0
  fi
  command -v "$BYXD_BIN" 2>/dev/null || true
}

show_byxd_info() {
  local byxd_path version
  byxd_path="$(resolve_byxd_bin)"
  [[ -n "$byxd_path" ]] || fail "missing command: $BYXD_BIN"

  ok "byxd bin: $BYXD_BIN"
  ok "byxd resolved path: $byxd_path"

  version="$("$BYXD_BIN" version --long 2>/dev/null || "$BYXD_BIN" version 2>/dev/null || true)"
  if [[ -n "$version" ]]; then
    echo "$version" | head -n 3 | sed 's/^/[OK] byxd version: /'
  else
    warn "could not read byxd version/build info"
  fi
}

check_amount_ubyx_cli_support() {
  local help
  help="$("$BYXD_BIN" tx payments create-payment-request --help 2>&1 || true)"
  if echo "$help" | grep -q '\[loja-id\] \[amount-ubyx\]'; then
    ok "create-payment-request supports positional amount-ubyx"
    if echo "$help" | grep -q -- '--amount-microbyx'; then
      warn "legacy flag alias still present: --amount-microbyx"
    fi
    return 0
  fi

  if echo "$help" | grep -q -- '--amount-microbyx'; then
    warn "legacy flag detected: --amount-microbyx"
  fi
  fail "byxd CLI does not support positional [amount-ubyx]. Rebuild local binary with: go build -o ./bin/byxd ./cmd/byxd and run with PATH=$(pwd)/bin:\$PATH"
}

check_e2e_script_uses_positional_amount() {
  local e2e_script="$ROOT_DIR/scripts/e2e_payments_webhook_ubyx.sh"
  local cmd_block

  cmd_block="$(awk '
    /tx payments create-payment-request[[:space:]]*\\$/ {in_cmd=1}
    in_cmd {
      print
      if ($0 !~ /\\[[:space:]]*$/) {
        exit
      }
    }
  ' "$e2e_script")"

  if [[ -z "$cmd_block" ]]; then
    fail "could not locate create-payment-request command block in e2e script"
  fi

  if echo "$cmd_block" | grep -q -- '--amount-ubyx'; then
    fail "e2e script still uses --amount-ubyx in executable create-payment-request command. Expected positional call: create-payment-request \"\$LOJA_ID\" \"\$AMOUNT_UBYX\""
  fi

  if echo "$cmd_block" | grep -q -- '--amount-microbyx'; then
    fail "e2e script still uses legacy --amount-microbyx in executable create-payment-request command"
  fi

  if echo "$cmd_block" | grep -Fq 'tx payments create-payment-request' \
    && echo "$cmd_block" | grep -Fq '"$LOJA_ID"' \
    && echo "$cmd_block" | grep -Fq '"$AMOUNT_UBYX"'; then
    ok "e2e script configured for positional amount-ubyx"
    return 0
  fi

  fail "could not confirm positional create-payment-request usage in executable command block"
}

check_create_payment_request_runtime_parse() {
  local -a probe_cmd
  local cmd_str
  local output
  local rc

  probe_cmd=(
    "$BYXD_BIN" tx payments create-payment-request "1" "1"
    --generate-only
    --from "$MERCHANT_KEY"
    --keyring-backend "$KEYRING_BACKEND"
    --account-number 0
    --sequence 0
    --fees 0ubyx
    --gas 200000
    --output json
  )

  cmd_str="$(printf '%q ' "${probe_cmd[@]}")"

  set +e
  output="$("${probe_cmd[@]}" 2>&1)"
  rc=$?
  set -e
  if echo "$output" | grep -q "accepts 0 arg(s), received 2"; then
    fail "byxd CLI help advertises positional amount-ubyx but runtime rejects positional args. Rebuild/regenerate/fix AutoCLI."
  fi

  if (( rc != 0 )); then
    warn "byxd positional parse probe failed (non-blocking).
probe command: $cmd_str
stderr: $(echo "$output" | head -n 2 | tr '\n' ' ')
hint: avoid --offline + --generate-only combinations that may conflict with local SDK/client config."
    return 0
  fi

  if echo "$output" | grep -qiE "unknown flag|required flag|expects [0-9]+ arg\\(s\\), received"; then
    warn "byxd positional parse probe failed (non-blocking).
probe command: $cmd_str
stderr: $(echo "$output" | head -n 2 | tr '\n' ' ')
hint: positional parse warning only; real validation continues in e2e execution."
    return 0
  fi

  ok "runtime positional parse probe accepted create-payment-request args"
}

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
  echo "BYXD_BIN=$BYXD_BIN"
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

  address="$("$BYXD_BIN" keys show "$key_name" -a --keyring-backend "$KEYRING_BACKEND" 2>/dev/null || true)"
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

need curl
need jq
need openssl
need node
need npm
show_byxd_info
check_amount_ubyx_cli_support
check_e2e_script_uses_positional_amount

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

check_create_payment_request_runtime_parse

echo "Preflight webhook/ubyx passed."
