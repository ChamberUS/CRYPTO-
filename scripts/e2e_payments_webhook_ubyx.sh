#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BYX_REST="${BYX_REST:-${REST:-http://127.0.0.1:1317}}"
BYX_RPC="${BYX_RPC:-http://127.0.0.1:26657}"
BYX_CHAIN_MODE="${BYX_CHAIN_MODE:-external}"
BYX_CHAIN_ID="${BYX_CHAIN_ID:-}"
LOJA_ID="${LOJA_ID:-1}"
AMOUNT_UBYX="${AMOUNT_UBYX:-500000}"
MERCHANT_KEY="${MERCHANT_KEY:-merchant}"
PAYER_KEY="${PAYER_KEY:-payer}"
KEYRING_BACKEND="${KEYRING_BACKEND:-test}"
STATE_PATH="${STATE_PATH:-$ROOT_DIR/webhook-relay/state.json}"
MOCK_MERCHANT_URL="${MOCK_MERCHANT_URL:-${MOCK_WEBHOOK_URL:-http://127.0.0.1:4000/webhook}}"
WEBHOOK_RELAY_URL="${WEBHOOK_RELAY_URL:-}"
MOCK_EVENTS_LOG_PATH="${MOCK_EVENTS_LOG_PATH:-/tmp/byx_mock_events.jsonl}"
MERCHANT_WEBHOOK_SECRET="${MERCHANT_WEBHOOK_SECRET:-devsecret}"
STRICT_WEBHOOK="${STRICT_WEBHOOK:-1}"
PAY_TIMEOUT_S="${PAY_TIMEOUT_S:-45}"
WEBHOOK_TIMEOUT_S="${WEBHOOK_TIMEOUT_S:-30}"
WAIT_TX_ATTEMPTS="${WAIT_TX_ATTEMPTS:-20}"

log() { echo "$*" >&2; }
die() { log "ERROR: $*"; exit 1; }

need() { command -v "$1" >/dev/null 2>&1 || die "missing command: $1"; }
need byxd
need curl
need jq
need openssl
need node

check_stack() {
  if ! curl -sf "$BYX_REST/cosmos/base/tendermint/v1beta1/syncing" >/dev/null; then
    log "REST unavailable: $BYX_REST"
    log "BYX_CHAIN_MODE atual: $BYX_CHAIN_MODE"
    log "Sugestao: BYX_CHAIN_MODE=external com chain ja ativa"
    log "Sugestao: BYX_CHAIN_MODE=byxd com byxd start --home <BYX_HOME>"
    log "Sugestao: BYX_CHAIN_MODE=custom com BYX_CHAIN_START_CMD=<comando>"
    log "Sugestao: BYX_CHAIN_MODE=ignite (pode depender de buf.build)"
    log "Sugestao: validar REST com: curl -sf $BYX_REST/cosmos/base/tendermint/v1beta1/syncing"
    log "Sugestao: executar diagnostico com: make doctor-webhook-ubyx"
    die "preflight stack failed (REST)"
  fi

  if ! curl -sf "$BYX_RPC/status" >/dev/null; then
    log "RPC unavailable: $BYX_RPC"
    log "BYX_CHAIN_MODE atual: $BYX_CHAIN_MODE"
    log "Sugestao: confirmar RPC com: curl -sf $BYX_RPC/status"
    log "Sugestao: executar diagnostico com: make doctor-webhook-ubyx"
    die "preflight stack failed (RPC)"
  fi
}

check_key() {
  local key=$1
  byxd keys show "$key" --keyring-backend "$KEYRING_BACKEND" >/dev/null 2>&1 || die "key '$key' not found"
}

tx_common_args() {
  local from="$1"
  local args=(--from "$from" --keyring-backend "$KEYRING_BACKEND" --broadcast-mode sync --output json --yes)
  if [[ -n "$BYX_CHAIN_ID" ]]; then
    args+=(--chain-id "$BYX_CHAIN_ID")
  fi
  if [[ -n "$BYX_RPC" ]]; then
    args+=(--node "$BYX_RPC")
  fi
  printf '%s\n' "${args[@]}"
}

wait_tx() {
  local txhash=$1
  local attempt=0
  local tmp
  tmp="$(mktemp)"
  while (( attempt < WAIT_TX_ATTEMPTS )); do
    if byxd query tx "$txhash" --output json --node "$BYX_RPC" >"$tmp" 2>/dev/null; then
      echo "$tmp"
      return 0
    fi
    attempt=$((attempt + 1))
    sleep 1
  done
  rm -f "$tmp"
  die "tx not indexed: $txhash"
}

create_request() {
  local out txhash txfile req_id
  out="$(mktemp)"

  mapfile -t tx_args < <(tx_common_args "$MERCHANT_KEY")
  byxd tx payments create-payment-request \
    --loja-id "$LOJA_ID" \
    --amount-ubyx "$AMOUNT_UBYX" \
    "${tx_args[@]}" >"$out" 2>&1 || {
      sed -n '1,200p' "$out" >&2
      rm -f "$out"
      die "create-payment-request failed"
    }

  txhash="$(jq -r '.txhash // .tx_response.txhash // empty' "$out")"
  [[ -n "$txhash" ]] || die "missing txhash in create-payment-request"
  txfile="$(wait_tx "$txhash")"

  req_id="$(jq -r '
    (.logs[]? | .events[]? | select(.type=="byx_payment_request_created") | .attributes[]? | select(.key=="request_id") | .value) //
    (.tx_response.logs[]? | .events[]? | select(.type=="byx_payment_request_created") | .attributes[]? | select(.key=="request_id") | .value) //
    empty
  ' "$txfile" | tail -n1)"

  if [[ -z "$req_id" ]]; then
    req_id="$(curl -s "$BYX_REST/byx/payments/v1/payment_requests/by_loja/$LOJA_ID?pagination.limit=1&pagination.reverse=true" | jq -r '.payment_requests[0].id // .paymentRequests[0].id // empty')"
  fi

  [[ -n "$req_id" ]] || die "could not resolve request_id"

  jq -e '.logs[]?.events[]? | select(.type=="byx_payment_request_created") | .attributes[]? | select(.key=="amount_ubyx")' "$txfile" >/dev/null || die "create event missing amount_ubyx"
  jq -e '.logs[]?.events[]? | .attributes[]? | select(.key=="amount_microbyx")' "$txfile" >/dev/null && die "found deprecated amount_microbyx in create event"

  rm -f "$out" "$txfile"
  echo "$req_id"
}

assert_query_fields() {
  local req_id=$1
  local qr pr
  qr="$(curl -sf "$BYX_REST/byx/payments/v1/payment_requests/$req_id/qr")"
  pr="$(curl -sf "$BYX_REST/byx/payments/v1/payment_requests/$req_id")"

  echo "$qr" | jq -e '.amount_ubyx' >/dev/null || die "QR response missing amount_ubyx"
  echo "$qr" | jq -e 'has("amount_microbyx")' >/dev/null && die "QR response still has amount_microbyx"
  echo "$pr" | jq -e '.payment_request.amount_ubyx // .paymentRequest.amount_ubyx' >/dev/null || die "payment request missing amount_ubyx"
  echo "$pr" | jq -e '.payment_request.amount_microbyx // .paymentRequest.amount_microbyx' >/dev/null && die "payment request still has amount_microbyx"
}

pay_request() {
  local req_id=$1
  mapfile -t tx_args < <(tx_common_args "$PAYER_KEY")
  byxd tx payments pay-payment-request \
    --request-id "$req_id" \
    "${tx_args[@]}" >/tmp/byx_pay_req.log 2>&1 || {
      sed -n '1,200p' /tmp/byx_pay_req.log >&2
      die "pay-payment-request failed"
    }

  local start status
  start="$(date +%s)"
  while true; do
    status="$(curl -s "$BYX_REST/byx/payments/v1/payment_requests/$req_id" | jq -r '.payment_request.status // .paymentRequest.status // empty')"
    [[ "$status" == "PAYMENT_STATUS_PAID" ]] && return 0
    (( $(date +%s) - start > PAY_TIMEOUT_S )) && die "timeout waiting PAID (last=$status)"
    sleep 1
  done
}

wait_webhook_state() {
  local req_id=$1
  local start event_id
  start="$(date +%s)"
  while true; do
    if [[ -f "$STATE_PATH" ]]; then
      event_id="$(jq -r --arg id "$req_id" '.sent[$id].eventId // empty' "$STATE_PATH")"
      if [[ -n "$event_id" ]]; then
        echo "$event_id"
        return 0
      fi
    fi
    (( $(date +%s) - start > WEBHOOK_TIMEOUT_S )) && break
    sleep 1
  done

  [[ "$STRICT_WEBHOOK" == "1" ]] && die "webhook state not observed for request $req_id"
  echo ""
}

send_duplicate_and_tampered() {
  local req_id=$1
  local event_id=$2
  local pr amount paid_at payload sig body code dup_resp

  pr="$(curl -sf "$BYX_REST/byx/payments/v1/payment_requests/$req_id")"
  amount="$(echo "$pr" | jq -r '.payment_request.amount_ubyx // .paymentRequest.amount_ubyx')"
  paid_at="$(echo "$pr" | jq -r '.payment_request.paid_at_unix // .paymentRequest.paid_at_unix // .paymentRequest.paidAtUnix // .payment_request.paidAtUnix')"

  payload="$(jq -nc \
    --argjson rid "$req_id" \
    --argjson lid "$LOJA_ID" \
    --argjson amt "$amount" \
    --argjson paid "$paid_at" \
    --arg ev "$event_id" \
    '{request_id:$rid, loja_id:$lid, amount_ubyx:$amt, paid_at_unix:$paid, event_id:$ev, trace_id:$ev}')"

  body="$payload"
  sig="$(printf '%s' "$body" | openssl dgst -sha256 -hmac "$MERCHANT_WEBHOOK_SECRET" -r | awk '{print $1}')"

  dup_resp="$(curl -s -o /tmp/byx_dup_resp.txt -w '%{http_code}' -X POST "$MOCK_MERCHANT_URL" \
    -H 'content-type: application/json' \
    -H "X-BYX-Signature: $sig" \
    -H "X-BYX-Idempotency-Key: $req_id" \
    -H "X-BYX-Event-Id: $event_id" \
    --data "$body")"
  [[ "$dup_resp" == "200" ]] || die "duplicate replay returned HTTP $dup_resp"

  code="$(curl -s -o /tmp/byx_bad_sig_resp.txt -w '%{http_code}' -X POST "$MOCK_MERCHANT_URL" \
    -H 'content-type: application/json' \
    -H 'X-BYX-Signature: bad-signature' \
    -H "X-BYX-Idempotency-Key: tampered-$req_id" \
    -H "X-BYX-Event-Id: tampered-$event_id" \
    --data "$body")"
  [[ "$code" == "401" ]] || die "tampered signature expected 401, got $code"

  if [[ -f "$MOCK_EVENTS_LOG_PATH" ]]; then
    jq -e 'select(.status=="ok" and .amount_ubyx != null)' "$MOCK_EVENTS_LOG_PATH" >/dev/null || die "mock log missing valid amount_ubyx event"
    jq -e 'select(.status=="duplicate")' "$MOCK_EVENTS_LOG_PATH" >/dev/null || die "mock log missing duplicate event"
    jq -e 'select(.status=="invalid_signature")' "$MOCK_EVENTS_LOG_PATH" >/dev/null || die "mock log missing invalid_signature event"
  fi

  echo "$payload"
}

main() {
  check_stack
  check_key "$MERCHANT_KEY"
  check_key "$PAYER_KEY"

  local req_id
  req_id="$(create_request)"
  log "REQUEST_ID=$req_id"

  assert_query_fields "$req_id"
  pay_request "$req_id"

  local event_id
  event_id="$(wait_webhook_state "$req_id")"
  [[ -n "$event_id" ]] && log "WEBHOOK_EVENT_ID=$event_id"

  local final_payload
  final_payload="$(send_duplicate_and_tampered "$req_id" "$event_id")"

  echo ""
  echo "E2E_UBYX_OK request_id=$req_id"
  echo "E2E_PAYMENT_PAYLOAD={\"loja_id\":$LOJA_ID,\"amount_ubyx\":$AMOUNT_UBYX}"
  echo "E2E_WEBHOOK_PAYLOAD=$final_payload"
  if [[ -n "$WEBHOOK_RELAY_URL" ]]; then
    echo "E2E_WEBHOOK_RELAY_URL=$WEBHOOK_RELAY_URL"
  fi
}

main "$@"
