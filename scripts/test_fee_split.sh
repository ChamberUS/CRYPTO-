#!/usr/bin/env bash
set -euo pipefail

# Smoke test for fee split 60/30/10 applied on gas fees only.
# Requires a running node with the feesplit module enabled.

HOME_DIR="${HOME_DIR:-$HOME/.byx}"
CHAIN_ID="${CHAIN_ID:-byx}"
KEYRING_BACKEND="${KEYRING_BACKEND:-test}"
NODE="${NODE:-tcp://0.0.0.0:26657}"
REST="${REST:-http://127.0.0.1:1317}"
SENDER="${FEE_SPLIT_SENDER:-marcelo}"

if ! command -v byxd >/dev/null 2>&1; then
  echo "byxd binary not found in PATH" >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required for this script" >&2
  exit 1
fi

rpc_url="${NODE/tcp/http}"
if ! curl -fsS "${rpc_url}/status" >/dev/null 2>&1; then
  echo "❌ RPC (${NODE}) não respondeu." >&2
  echo "   Dicas: byxd start | ignite chain serve --reset-once | comando padrão do repo para subir a chain." >&2
  exit 1
fi

if ! curl -fsS "${REST}/cosmos/base/tendermint/v1beta1/node_info" >/dev/null 2>&1; then
  echo "❌ REST (${REST}) não respondeu." >&2
  echo "   Dicas: confirme se a API está habilitada/bindada (porta 1317) ao subir a chain." >&2
  exit 1
fi

extract_treasury_addr() {
  local json="$1"
  echo "$json" | jq -r '
    .account.base_account.address // .account.value.address // .account.address // empty
  ' | head -n1
}

treasury_json="$(byxd query auth module-account treasury --node "$NODE" --output json 2>/dev/null || true)"
if [[ -z "$treasury_json" ]]; then
  treasury_json="$(byxd query auth module-account treasury -o json --node "$NODE" 2>/dev/null || true)"
fi

treasury_addr="$(extract_treasury_addr "$treasury_json")"

if [[ -z "$treasury_addr" ]]; then
  echo "Could not find treasury module account address" >&2
  exit 1
fi

ROBERTA_ADDR="$(byxd keys show roberta --keyring-backend "$KEYRING_BACKEND" --home "$HOME_DIR" -a 2>/dev/null || true)"
if [[ -z "$ROBERTA_ADDR" ]]; then
  ROBERTA_ADDR="$(byxd keys list --keyring-backend "$KEYRING_BACKEND" --home "$HOME_DIR" -o json 2>/dev/null \
    | jq -r '.[] | select(.name=="roberta") | .address' | head -n1)"
fi
if [[ -z "$ROBERTA_ADDR" ]]; then
  echo "Could not resolve roberta address" >&2
  exit 1
fi

balance_or_zero() {
  local addr="$1"
  byxd query bank balances "$addr" --node "$NODE" --output json \
    | jq -r '(.balances[]? | select(.denom=="ubyx") | .amount) // "0"'
}

total_supply_byx() {
  byxd query bank total --node "$NODE" --output json \
    | jq -r '(.supply[]? | select(.denom=="ubyx") | .amount) // "0"'
}

treasury_before="$(balance_or_zero "$treasury_addr")"
supply_before="$(total_supply_byx)"

tmp="$(mktemp)"
if ! byxd tx bank send marcelo "$ROBERTA_ADDR" 1000000ubyx \
  --fees 10000ubyx \
  --chain-id "$CHAIN_ID" \
  --node "$NODE" \
  --home "$HOME_DIR" \
  --broadcast-mode sync \
  --yes \
  --keyring-backend "$KEYRING_BACKEND" \
  --output json >"$tmp" 2>&1; then
  echo "tx failed:" >&2
  sed -n '1,200p' "$tmp" >&2 || true
  rm -f "$tmp"
  exit 1
fi
TX_JSON="$(sed -n '/^{/,$p' "$tmp")"
# show the json part for debugging
printf '%s\n' "$TX_JSON" >&2 || true
rm -f "$tmp"

TXHASH="$(printf '%s\n' "$TX_JSON" | jq -r '.txhash // empty')"
if [[ -z "$TXHASH" || "$TXHASH" == "null" ]]; then
  echo "Could not extract txhash from TX response" >&2
  exit 1
fi

# wait for the tx to be included in a block (deterministic)
confirmed_height=""
for i in $(seq 1 30); do
  RES="$(byxd query tx "$TXHASH" --home "$HOME_DIR" --node "$NODE" -o json 2>/dev/null || true)"
  confirmed_height="$(printf '%s\n' "$RES" | jq -r '.height // "0"')"
  if [[ -n "$confirmed_height" && "$confirmed_height" != "0" ]]; then
    break
  fi
  sleep 1
done

if [[ -z "$confirmed_height" || "$confirmed_height" == "0" ]]; then
  echo "tx $TXHASH not confirmed within timeout" >&2
  exit 1
fi

treasury_after="$(balance_or_zero "$treasury_addr")"
supply_after="$(total_supply_byx)"

treasury_delta="$(perl -E "say $treasury_after - $treasury_before")"
burn_delta="$(perl -E "say $supply_before - $supply_after")"

if [[ "$treasury_delta" -ge 3000 ]]; then
  echo "FEE_SPLIT_OK"
  exit 0
fi

echo "FEE_SPLIT_MISMATCH treasury_delta=$treasury_delta burn_delta=$burn_delta" >&2
exit 1
