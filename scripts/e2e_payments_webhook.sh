#!/usr/bin/env bash
set -euo pipefail

# E2E helper to create + pay a payment request and wait until status=PAID.

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
REST=${REST:-http://127.0.0.1:1317}
LOJA_ID=${LOJA_ID:-1}
AMOUNT_MICROBYX=${AMOUNT_MICROBYX:-500000} # 0.5 BYX default
MERCHANT_KEY=${MERCHANT_KEY:-merchant}
PAYER_KEY=${PAYER_KEY:-payer}
KEYRING_BACKEND=${KEYRING_BACKEND:-test}
POLL_SECONDS=${POLL_SECONDS:-15}
PAY_POLL_TIMEOUT_S=${PAY_POLL_TIMEOUT_S:-30}
STATE_PATH=${STATE_PATH:-"$ROOT_DIR/webhook-relay/state.json"}
STRICT_WEBHOOK=${STRICT_WEBHOOK:-0}
DEBUG_E2E=${DEBUG_E2E:-0}
KEEP_TMP=${KEEP_TMP:-0}
WAIT_TX_ATTEMPTS=${WAIT_TX_ATTEMPTS:-20}

timestamp() {
  date +"%Y-%m-%dT%H:%M:%S%z"
}

log() {
  if [[ "${DEBUG_E2E}" == "1" ]]; then
    echo "[$(timestamp)] $*" >&2
  else
    echo "$*" >&2
  fi
}

die() {
  log "ERROR: $*"
  exit 1
}

if [[ "${DEBUG_E2E}" == "1" ]]; then
  set -x
  PS4='+ [$(date +"%H:%M:%S")] '
fi

command -v byxd >/dev/null 2>&1 || die "byxd CLI is required on PATH"
command -v curl >/dev/null 2>&1 || die "curl is required"
command -v jq >/dev/null 2>&1 || die "jq is required"

echo "REST=${REST}"
echo "LOJA_ID=${LOJA_ID}"
echo "AMOUNT_MICROBYX=${AMOUNT_MICROBYX}"
echo "MERCHANT_KEY=${MERCHANT_KEY}"
echo "PAYER_KEY=${PAYER_KEY}"
echo "KEYRING_BACKEND=${KEYRING_BACKEND}"
echo "STATE_PATH=${STATE_PATH}"
echo "STRICT_WEBHOOK=${STRICT_WEBHOOK}"
echo "DEBUG_E2E=${DEBUG_E2E}"

check_key() {
  local key=$1
  if ! byxd keys show "${key}" --keyring-backend "${KEYRING_BACKEND}" >/dev/null 2>&1; then
    die "Key '${key}' not found. Hint: set KEYRING_BACKEND (default 'test') and import/create keys (byxd keys list)."
  fi
}

wait_tx() {
  local txhash=$1
  local attempt=0
  local tmp
  tmp=$(mktemp)
  while (( attempt < WAIT_TX_ATTEMPTS )); do
    if byxd query tx "${txhash}" --output json >"$tmp" 2>/dev/null; then
      echo "$tmp"
      return 0
    fi
    attempt=$((attempt + 1))
    log "aguardando tx indexar... tentativa ${attempt}/${WAIT_TX_ATTEMPTS}" >&2
    sleep 1
  done
  [[ "${KEEP_TMP}" == "1" ]] || rm -f "$tmp"
  die "Failed to query tx ${txhash} after ${WAIT_TX_ATTEMPTS} attempts"
}

create_payment_request() {
  log "Creating payment request..."
  local tmp
  tmp=$(mktemp)
  if ! byxd tx payments create-payment-request \
    --loja-id "${LOJA_ID}" \
    --amount-microbyx "${AMOUNT_MICROBYX}" \
    --from "${MERCHANT_KEY}" \
    --keyring-backend "${KEYRING_BACKEND}" \
    --broadcast-mode sync \
    --output json \
    --yes >"$tmp" 2>&1; then
    log "create-payment-request failed, output:"
    cat "$tmp" >&2
    [[ "${KEEP_TMP}" == "1" ]] || rm -f "$tmp"
    die "create-payment-request failed"
  fi

  local code
  code=$(jq -r '.code // .tx_response.code // empty' "$tmp")
  if [[ -n "${code}" && "${code}" != "0" ]]; then
    local raw
    raw=$(jq -r '.raw_log // .tx_response.raw_log // empty' "$tmp")
    log "TX error code=${code} raw_log=${raw}"
    [[ "${KEEP_TMP}" == "1" ]] || rm -f "$tmp"
    die "tx failed code=${code}"
  fi

  local txhash
  txhash=$(jq -r '.txhash // .tx_response.txhash // empty' "$tmp")
  echo "TXHASH=${txhash}" >&2
  if [[ -z "${txhash}" || "${txhash}" == "null" ]]; then
    [[ "${KEEP_TMP}" == "1" ]] || rm -f "$tmp"
    die "Could not extract txhash from create-payment-request"
  fi

  local txfile
  txfile=$(wait_tx "${txhash}")
  [[ "${KEEP_TMP}" == "1" ]] || rm -f "$tmp"

  local req_id
  req_id=$(jq -r '
    (.logs[]? | .events[]? | select(.type=="byx_payment_request_created") | .attributes[]? | select(.key=="request_id") | .value) //
    (.tx_response.logs[]? | .events[]? | select(.type=="byx_payment_request_created") | .attributes[]? | select(.key=="request_id") | .value) //
    (.logs[]? | .events[]? | .attributes[]? | select(.key=="request_id") | .value) //
    (.tx_response.logs[]? | .events[]? | .attributes[]? | select(.key=="request_id") | .value) //
    (.events[]? | select(.type=="byx_payment_request_created") | .attributes[]? | select(.key=="request_id") | .value) //
    empty
  ' "$txfile" | tail -n1)

  if [[ -z "${req_id}" || "${req_id}" == "null" ]]; then
    req_id=$(jq -r '..|objects|.request_id? // empty' "$txfile" | head -n1)
  fi
  [[ "${KEEP_TMP}" == "1" ]] || rm -f "$txfile"

  if [[ -z "${req_id}" || "${req_id}" == "null" ]]; then
    for attempt in {1..10}; do
      log "Could not parse request_id from tx result, retrying REST query (${attempt}/10)..."
      req_id=$(curl -s "${REST}/byx/payments/v1/payment_requests/by_loja/${LOJA_ID}?pagination.limit=1&pagination.reverse=true" \
        | jq -r '.payment_requests[0].id // .paymentRequests[0].id')
      if [[ -n "${req_id}" && "${req_id}" != "null" ]]; then
        break
      fi
      sleep 1
    done
  fi

  if [[ -z "${req_id}" || "${req_id}" == "null" ]]; then
    die "Failed to obtain request_id."
  fi

  echo "${req_id}"
}

pay_request() {
  local req_id="$1"
  log "Paying request ${req_id} via CLI..."

  # Dispara a tx, mas nao confia no JSON do CLI (gas estimate pode poluir stdout)
  local tmp
  tmp="$(mktemp)"
  if ! byxd tx payments pay-payment-request \
    --request-id "${req_id}" \
    --from "${PAYER_KEY}" \
    --keyring-backend "${KEYRING_BACKEND}" \
    --broadcast-mode sync \
    --yes >"${tmp}" 2>&1; then
    log "ERROR: pay-payment-request tx failed (raw output):"
    sed -n "1,200p" "${tmp}" 1>&2 || true
    rm -f "${tmp}"
    die "pay-payment-request failed"
  fi
  rm -f "${tmp}"

  # Confirma por REST ate virar PAID (idempotente e a prova de ruido)
  log "Polling status until PAID (timeout ${PAY_POLL_TIMEOUT_S}s)..."
  local start now status
  start="$(date +%s)"
  while true; do
    status="$(curl -s "${REST}/byx/payments/v1/payment_requests/${req_id}" | jq -r '.payment_request.status // .paymentRequest.status // empty')"
    if [[ "${status}" == "PAYMENT_STATUS_PAID" ]]; then
      echo "STATUS=PAID (request ${req_id})"
      return 0
    fi
    now="$(date +%s)"
    if (( now - start > PAY_POLL_TIMEOUT_S )); then
      log "ERROR: timeout waiting PAID. last_status=${status}"
      die "timeout waiting PAID"
    fi
    sleep 1
  done
}

wait_paid() {
  local req_id=$1
  log "Polling status until PAID (timeout ${POLL_SECONDS}s)..."
  local elapsed=0
  local last_status=""
  local paid_at=""
  while (( elapsed < POLL_SECONDS )); do
    local status
    status=$(curl -s "${REST}/byx/payments/v1/payment_requests/${req_id}" \
      | jq -r '.payment_request.status // .paymentRequest.status // ""')
    last_status="$status"

    case "${status}" in
      *PAID*|2)
        paid_at=$(curl -s "${REST}/byx/payments/v1/payment_requests/${req_id}" \
          | jq -r '.payment_request.paid_at_unix // .paymentRequest.paid_at_unix // .paymentRequest.paidAtUnix // .payment_request.paidAtUnix // empty')
        echo "STATUS=PAID (request ${req_id})"
        PAID_AT_UNIX="${paid_at}"
        return 0
        ;;
      *)
        ;;
    esac

    sleep 1
    elapsed=$((elapsed + 1))
  done

  die "Timed out waiting for request ${req_id} to become PAID (last status='${last_status}')"
}

validate_state() {
  local req_id=$1
  local paid_at=${2:-}
  log "Validating webhook state..."
  local path="${STATE_PATH}"
  if [[ ! -f "${path}" ]]; then
    log "WEBHOOK_STATE=WARN file not found at ${path}. Start relay and mock merchant? (STRICT_WEBHOOK=${STRICT_WEBHOOK})"
    [[ "${STRICT_WEBHOOK}" == "1" ]] && die "Strict mode: missing state file."
    return
  fi

  local attempts=0
  local found=0
  local event_id=""
  if [[ -n "${paid_at}" && "${paid_at}" != "null" ]]; then
    event_id="${req_id}:${paid_at}"
  fi
  while (( attempts < 15 )); do
    if [[ -n "${event_id}" ]]; then
      found=$(jq -e --arg id "${req_id}" --arg ev "${event_id}" '
        (.sent[$id].eventId == $ev) or (.failures[$id].eventId == $ev)
      ' "${path}" >/dev/null 2>&1; echo $?)
    else
      found=$(jq -e --arg id "${req_id}" '
        (.sent[$id]) or (.failures[$id])
      ' "${path}" >/dev/null 2>&1; echo $?)
    fi
    if [[ "${found}" == "0" ]]; then
      echo "WEBHOOK_STATE=OK (request ${req_id}${event_id:+ event_id=${event_id}})"
      return
    fi
    sleep 1
    attempts=$((attempts + 1))
  done

  log "WEBHOOK_STATE=WARN not found in state after ${attempts}s (request ${req_id}${event_id:+ event_id=${event_id}})"
  [[ "${STRICT_WEBHOOK}" == "1" ]] && die "Strict mode: webhook state missing."
}

check_key "${MERCHANT_KEY}"
check_key "${PAYER_KEY}"

REQUEST_ID=$(create_payment_request)
echo "REQUEST_ID=${REQUEST_ID}"

pay_request "${REQUEST_ID}"
wait_paid "${REQUEST_ID}"

validate_state "${REQUEST_ID}" "${PAID_AT_UNIX:-}"

echo ""
echo "E2E OK: payment PAID, webhook relay should have sent notification"
