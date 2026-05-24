#!/usr/bin/env bash
set -euo pipefail

# Simple dedupe validation script (Sprint 9.2)

REST="${REST_ENDPOINT:-http://127.0.0.1:1317}"
MERCHANT_KEY="${MERCHANT_KEY:-marcelo}"
KEYRING_BACKEND="${KEYRING_BACKEND:-test}"
LOJA_ID="${LOJA_ID:-1}"
AMOUNT_UBYX="${AMOUNT_UBYX:-500000}"

echo "Using REST=${REST}"

create() {
  local memo="$1"
  byxd tx payments create-payment-request \
    --loja-id "${LOJA_ID}" \
    --amount-ubyx "${AMOUNT_UBYX}" \
    --memo "${memo}" \
    --from "${MERCHANT_KEY}" \
    --keyring-backend "${KEYRING_BACKEND}" \
    --broadcast-mode sync \
    --yes \
    --output json | jq -r '.id // .tx_response.logs[]?.events[]? | select(.type=="byx_payment_request_created") | .attributes[]? | select(.key=="request_id") | .value' | tail -n1
}

id1="$(create "A")"
echo "First request id=${id1}"

id1b="$(create "A")"
echo "Second request (same memo) id=${id1b}"

if [[ "${id1}" != "${id1b}" ]]; then
  echo "❌ dedupe failed: expected same id for same fingerprint"
  exit 1
fi

id2="$(create "B")"
echo "Third request (memo B) id=${id2}"

if [[ "${id2}" == "${id1}" ]]; then
  echo "❌ dedupe failed: expected new id for different memo"
  exit 1
fi

echo "✅ dedupe script OK"
