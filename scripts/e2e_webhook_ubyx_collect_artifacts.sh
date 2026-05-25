#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
E2E_DIR="${E2E_DIR:-$ROOT_DIR/.e2e/webhook-ubyx}"
STATE_PATH="${STATE_PATH:-$E2E_DIR/state.json}"
MOCK_EVENTS_LOG_PATH="${MOCK_EVENTS_LOG_PATH:-$E2E_DIR/mock-events.jsonl}"

mkdir -p "$E2E_DIR"

copy_if_exists() {
  local src=$1
  local dst=$2
  if [[ "$src" == "$dst" ]]; then
    return 0
  fi
  if [[ -e "$src" && -e "$dst" ]] && [[ "$src" -ef "$dst" ]]; then
    return 0
  fi
  if [[ -f "$src" ]]; then
    cp "$src" "$dst"
  fi
}

copy_if_exists "$STATE_PATH" "$E2E_DIR/state.json"
copy_if_exists "$MOCK_EVENTS_LOG_PATH" "$E2E_DIR/mock-events.jsonl"

cat >"$E2E_DIR/artifacts.txt" <<ART
Artifacts collected at: $(date -u +%Y-%m-%dT%H:%M:%SZ)
- chain log: $E2E_DIR/chain.log
- mock merchant log: $E2E_DIR/mock-merchant.log
- webhook relay log: $E2E_DIR/webhook-relay.log
- preflight log: $E2E_DIR/preflight.log
- doctor log: $E2E_DIR/doctor.log
- e2e log: $E2E_DIR/e2e.log
- state file: $E2E_DIR/state.json
- mock events: $E2E_DIR/mock-events.jsonl
- create-payment-request broadcast: $E2E_DIR/create_payment_request_broadcast.json
- create-payment-request broadcast stderr: $E2E_DIR/create_payment_request_broadcast.stderr.txt
- create-payment-request broadcast raw: $E2E_DIR/create_payment_request_broadcast.raw.txt
- create-payment-request tx: $E2E_DIR/create_payment_request_tx.json
- create-payment-request txhash: $E2E_DIR/txhash_create_payment_request.txt
- create-payment-request command: $E2E_DIR/create_payment_request_command.txt
- e2e memo: $E2E_DIR/e2e_memo.txt
- payment request query json: $E2E_DIR/payment_request_query.json
- payment request query http: $E2E_DIR/payment_request_query.http.txt
- payment request qr json: $E2E_DIR/payment_request_qr.json
- payment request qr http: $E2E_DIR/payment_request_qr.http.txt
- merchant query by id: $E2E_DIR/merchant_query_by_id.json
- merchant query all: $E2E_DIR/merchant_query_all.json
- create-merchant broadcast: $E2E_DIR/create_merchant_broadcast.json
- create-merchant broadcast stderr: $E2E_DIR/create_merchant_broadcast.stderr.txt
- create-merchant broadcast raw: $E2E_DIR/create_merchant_broadcast.raw.txt
- create-merchant tx: $E2E_DIR/create_merchant_tx.json
- pay-request broadcast: $E2E_DIR/pay_request_broadcast.json
- pay-request broadcast stderr: $E2E_DIR/pay_request_broadcast.stderr.txt
- pay-request broadcast raw: $E2E_DIR/pay_request_broadcast.raw.txt
- pay-request tx: $E2E_DIR/pay_request_tx.json
- pay-request command: $E2E_DIR/pay_request_command.txt
- merchant id: $E2E_DIR/merchant_id.txt
- resolved request id: $E2E_DIR/request_id.txt
- wait-tx last response: $E2E_DIR/wait_tx_last_response.txt
- merchant signer info: $E2E_DIR/merchant_signer_info.txt
- merchant account on-chain: $E2E_DIR/merchant_account_onchain.json
- chain status: $E2E_DIR/chain_status.json
- chain mode: $E2E_DIR/chain_mode.txt
- env summary: $E2E_DIR/env_summary.txt
- startup command (masked): $E2E_DIR/startup_command.txt
- failure reason: $E2E_DIR/failure_reason.txt
ART

echo "[collect] artifacts directory: $E2E_DIR"
