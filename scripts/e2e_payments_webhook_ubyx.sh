#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BYX_REST="${BYX_REST:-${REST:-http://127.0.0.1:1317}}"
BYX_RPC="${BYX_RPC:-http://127.0.0.1:26657}"
BYX_CHAIN_MODE="${BYX_CHAIN_MODE:-external}"
BYX_CHAIN_ID="${BYX_CHAIN_ID:-}"
BYXD_BIN="${BYXD_BIN:-byxd}"
LOJA_ID="${LOJA_ID:-1}"
AMOUNT_UBYX="${AMOUNT_UBYX:-500000}"
MERCHANT_KEY="${MERCHANT_KEY:-merchant}"
PAYER_KEY="${PAYER_KEY:-payer}"
KEYRING_BACKEND="${KEYRING_BACKEND:-test}"
STATE_PATH="${STATE_PATH:-$ROOT_DIR/.e2e/webhook-ubyx/state.json}"
MOCK_MERCHANT_URL="${MOCK_MERCHANT_URL:-${MOCK_WEBHOOK_URL:-http://127.0.0.1:4000/webhook}}"
WEBHOOK_RELAY_URL="${WEBHOOK_RELAY_URL:-}"
MOCK_EVENTS_LOG_PATH="${MOCK_EVENTS_LOG_PATH:-/tmp/byx_mock_events.jsonl}"
MERCHANT_WEBHOOK_SECRET="${MERCHANT_WEBHOOK_SECRET:-devsecret}"
STRICT_WEBHOOK="${STRICT_WEBHOOK:-1}"
PAY_TIMEOUT_S="${PAY_TIMEOUT_S:-45}"
WEBHOOK_TIMEOUT_S="${WEBHOOK_TIMEOUT_S:-30}"
WAIT_TX_ATTEMPTS="${WAIT_TX_ATTEMPTS:-60}"
WAIT_TX_SLEEP_S="${WAIT_TX_SLEEP_S:-2}"
E2E_DIR="${E2E_DIR:-$ROOT_DIR/.e2e/webhook-ubyx}"
CREATE_REQUEST_BROADCAST_JSON="$E2E_DIR/create_payment_request_broadcast.json"
CREATE_REQUEST_BROADCAST_STDERR_TXT="$E2E_DIR/create_payment_request_broadcast.stderr.txt"
CREATE_REQUEST_BROADCAST_RAW_TXT="$E2E_DIR/create_payment_request_broadcast.raw.txt"
CREATE_REQUEST_TX_JSON="$E2E_DIR/create_payment_request_tx.json"
TXHASH_CREATE_REQUEST_TXT="$E2E_DIR/txhash_create_payment_request.txt"
REQUEST_ID_TXT="$E2E_DIR/request_id.txt"
WAIT_TX_LAST_RESPONSE_TXT="$E2E_DIR/wait_tx_last_response.txt"
CREATE_REQUEST_COMMAND_TXT="$E2E_DIR/create_payment_request_command.txt"
MERCHANT_ACCOUNT_ONCHAIN_JSON="$E2E_DIR/merchant_account_onchain.json"
MERCHANT_SIGNER_INFO_TXT="$E2E_DIR/merchant_signer_info.txt"
CHAIN_STATUS_JSON="$E2E_DIR/chain_status.json"
MERCHANT_QUERY_BY_ID_JSON="$E2E_DIR/merchant_query_by_id.json"
MERCHANT_QUERY_ALL_JSON="$E2E_DIR/merchant_query_all.json"
MERCHANT_CREATE_BROADCAST_JSON="$E2E_DIR/create_merchant_broadcast.json"
MERCHANT_CREATE_BROADCAST_STDERR_TXT="$E2E_DIR/create_merchant_broadcast.stderr.txt"
MERCHANT_CREATE_BROADCAST_RAW_TXT="$E2E_DIR/create_merchant_broadcast.raw.txt"
MERCHANT_CREATE_TX_JSON="$E2E_DIR/create_merchant_tx.json"
MERCHANT_ID_TXT="$E2E_DIR/merchant_id.txt"
PAY_REQUEST_BROADCAST_JSON="$E2E_DIR/pay_request_broadcast.json"
PAY_REQUEST_BROADCAST_STDERR_TXT="$E2E_DIR/pay_request_broadcast.stderr.txt"
PAY_REQUEST_BROADCAST_RAW_TXT="$E2E_DIR/pay_request_broadcast.raw.txt"

log() { echo "$*" >&2; }
die() { log "ERROR: $*"; exit 1; }

mkdir -p "$E2E_DIR"

need() { command -v "$1" >/dev/null 2>&1 || die "missing command: $1"; }
need_byxd_bin() {
  if [[ "$BYXD_BIN" == */* ]]; then
    [[ -x "$BYXD_BIN" ]] || die "BYXD_BIN is not executable: $BYXD_BIN"
    return 0
  fi
  command -v "$BYXD_BIN" >/dev/null 2>&1 || die "missing command: $BYXD_BIN"
}

check_byxd_amount_ubyx_support() {
  local help
  help="$("$BYXD_BIN" tx payments create-payment-request --help 2>&1 || true)"
  if echo "$help" | grep -q '\[loja-id\] \[amount-ubyx\]'; then
    if echo "$help" | grep -q -- '--amount-microbyx'; then
      log "Detected positional amount-ubyx with legacy flag alias --amount-microbyx in: $BYXD_BIN"
    fi
    return 0
  fi

  if echo "$help" | grep -q -- '--amount-microbyx'; then
    log "Detected legacy CLI flag --amount-microbyx in: $BYXD_BIN"
  fi
  die "byxd CLI does not expose positional [amount-ubyx]. Rebuild local binary with: go build -o ./bin/byxd ./cmd/byxd and run with PATH=$(pwd)/bin:\$PATH"
}

need_byxd_bin
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
  "$BYXD_BIN" keys show "$key" --keyring-backend "$KEYRING_BACKEND" >/dev/null 2>&1 || die "key '$key' not found"
}

resolve_chain_id() {
  if [[ -n "$BYX_CHAIN_ID" ]]; then
    printf '%s\n' "$BYX_CHAIN_ID"
    return 0
  fi
  curl -sf "$BYX_RPC/status" >"$CHAIN_STATUS_JSON" || die "failed to query chain status from $BYX_RPC"
  BYX_CHAIN_ID="$(jq -r '.result.node_info.network // empty' "$CHAIN_STATUS_JSON")"
  [[ -n "$BYX_CHAIN_ID" ]] || die "failed to resolve chain-id from $BYX_RPC/status"
  printf '%s\n' "$BYX_CHAIN_ID"
}

tx_common_args() {
  local from="$1"
  local chain_id
  chain_id="$(resolve_chain_id)"
  local args=(
    --from "$from"
    --keyring-backend "$KEYRING_BACKEND"
    --chain-id "$chain_id"
    --node "$BYX_RPC"
    --fees 0ubyx
    --gas auto
    --gas-adjustment 1.3
    --broadcast-mode sync
    --output json
    --yes
  )
  printf '%s\n' "${args[@]}"
}

log_sanitized_create_request_command() {
  local chain_id=$1
  cat >"$CREATE_REQUEST_COMMAND_TXT" <<EOF
BYXD_BIN=$BYXD_BIN
loja_id=$LOJA_ID
amount_ubyx=$AMOUNT_UBYX
from=$MERCHANT_KEY
chain_id=$chain_id
node=$BYX_RPC
fees=0ubyx
gas=auto
gas_adjustment=1.3
keyring_backend=$KEYRING_BACKEND
command=$BYXD_BIN tx payments create-payment-request $LOJA_ID $AMOUNT_UBYX --from $MERCHANT_KEY --keyring-backend $KEYRING_BACKEND --chain-id $chain_id --node $BYX_RPC --fees 0ubyx --gas auto --gas-adjustment 1.3 --broadcast-mode sync --yes --output json
EOF
  sed 's/^/CREATE_REQUEST_CMD /' "$CREATE_REQUEST_COMMAND_TXT" >&2
}

extract_account_number_from_file() {
  local file=$1
  jq -r '
    .. | objects | .account_number? // empty
  ' "$file" | head -n1
}

extract_sequence_from_file() {
  local file=$1
  jq -r '
    .. | objects | .sequence? // empty
  ' "$file" | head -n1
}

log_merchant_onchain_diagnostics() {
  local merchant_addr chain_id account_number sequence
  merchant_addr="$("$BYXD_BIN" keys show "$MERCHANT_KEY" -a --keyring-backend "$KEYRING_BACKEND")"
  chain_id="$(resolve_chain_id)"
  printf 'merchant_key=%s\nmerchant_address=%s\nchain_id=%s\n' "$MERCHANT_KEY" "$merchant_addr" "$chain_id" >"$MERCHANT_SIGNER_INFO_TXT"
  "$BYXD_BIN" q auth account "$merchant_addr" --node "$BYX_RPC" -o json >"$MERCHANT_ACCOUNT_ONCHAIN_JSON" 2>/dev/null \
    || curl -sf "$BYX_REST/cosmos/auth/v1beta1/accounts/$merchant_addr" >"$MERCHANT_ACCOUNT_ONCHAIN_JSON" \
    || die "failed to query merchant on-chain account diagnostics"
  require_json_file "$MERCHANT_ACCOUNT_ONCHAIN_JSON" "merchant on-chain account diagnostics"
  account_number="$(extract_account_number_from_file "$MERCHANT_ACCOUNT_ONCHAIN_JSON")"
  sequence="$(extract_sequence_from_file "$MERCHANT_ACCOUNT_ONCHAIN_JSON")"
  {
    printf 'account_number=%s\n' "${account_number:-}"
    printf 'sequence=%s\n' "${sequence:-}"
  } >>"$MERCHANT_SIGNER_INFO_TXT"
  sed 's/^/MERCHANT_SIGNER /' "$MERCHANT_SIGNER_INFO_TXT" >&2
}

validate_real_tx_args() {
  local arg
  for arg in "$@"; do
    case "$arg" in
      --account-number|--sequence|--offline|--generate-only)
        die "real tx contains forbidden flag: $arg"
        ;;
    esac
  done
}

read_broadcast_field() {
  local file=$1
  local expr=$2
  jq -r "$expr" "$file"
}

fail_if_broadcast_rejected() {
  local file=$1
  local context=$2
  local code codespace raw_log
  require_json_file "$file" "$context broadcast"
  code="$(read_broadcast_field "$file" '.code // .tx_response.code // 0')"
  codespace="$(read_broadcast_field "$file" '.codespace // .tx_response.codespace // empty')"
  raw_log="$(read_broadcast_field "$file" '.raw_log // .tx_response.raw_log // empty')"
  if [[ "$code" != "0" ]]; then
    die "$context broadcast rejected: code=$code codespace=${codespace:-<empty>} raw_log=$raw_log"
  fi
}

dump_stdout_stderr_for_debug() {
  local stdout_file=$1
  local stderr_file=$2
  [[ -f "$stdout_file" ]] && [[ -s "$stdout_file" ]] && {
    log "stdout ($stdout_file):"
    sed -n '1,200p' "$stdout_file" >&2
  }
  [[ -f "$stderr_file" ]] && [[ -s "$stderr_file" ]] && {
    log "stderr ($stderr_file):"
    sed -n '1,200p' "$stderr_file" >&2
  }
}

extract_json_from_raw() {
  local raw_file=$1
  local json_file=$2
  local candidate_file
  candidate_file="$(mktemp)"

  if jq -e . "$raw_file" >/dev/null 2>&1; then
    cp "$raw_file" "$json_file"
    rm -f "$candidate_file"
    return 0
  fi

  grep -E '^[[:space:]]*\{.*\}[[:space:]]*$' "$raw_file" | tail -n1 >"$candidate_file" || true
  if [[ -s "$candidate_file" ]] && jq -e . "$candidate_file" >/dev/null 2>&1; then
    cp "$candidate_file" "$json_file"
    rm -f "$candidate_file"
    return 0
  fi

  rm -f "$candidate_file"
  return 1
}

run_tx_json() {
  local label=$1
  local json_file=$2
  local stderr_file=$3
  local raw_file=$4
  shift 4

  local raw_tmp stderr_tmp txhash
  raw_tmp="$(mktemp)"
  stderr_tmp="$(mktemp)"

  if [[ "${1:-}" == "--" ]]; then
    shift
  fi

  "$@" >"$raw_tmp" 2>"$stderr_tmp" || {
    cp "$raw_tmp" "$raw_file" 2>/dev/null || true
    cp "$stderr_tmp" "$stderr_file" 2>/dev/null || true
    dump_stdout_stderr_for_debug "$raw_tmp" "$stderr_tmp"
    rm -f "$raw_tmp" "$stderr_tmp"
    die "$label failed"
  }

  cp "$raw_tmp" "$raw_file"
  cp "$stderr_tmp" "$stderr_file"
  if ! extract_json_from_raw "$raw_tmp" "$json_file"; then
    dump_stdout_stderr_for_debug "$raw_file" "$stderr_file"
    rm -f "$raw_tmp" "$stderr_tmp"
    die "$label returned invalid json"
  fi

  fail_if_broadcast_rejected "$json_file" "$label"
  txhash="$(read_broadcast_field "$json_file" '.txhash // .tx_response.txhash // empty')"
  [[ -n "$txhash" ]] || {
    dump_stdout_stderr_for_debug "$raw_file" "$stderr_file"
    rm -f "$raw_tmp" "$stderr_tmp"
    die "$label broadcast missing txhash"
  }

  rm -f "$raw_tmp" "$stderr_tmp"
  printf '%s\n' "$txhash"
}

query_merchant_by_id() {
  local merchant_id=$1
  if "$BYXD_BIN" query lojas merchant "$merchant_id" --node "$BYX_RPC" -o json >"$MERCHANT_QUERY_BY_ID_JSON" 2>/dev/null; then
    return 0
  fi
  curl -sf "$BYX_REST/byx/lojas/v1/merchant/$merchant_id" >"$MERCHANT_QUERY_BY_ID_JSON"
}

query_all_merchants() {
  if "$BYXD_BIN" query lojas merchant-all --node "$BYX_RPC" -o json >"$MERCHANT_QUERY_ALL_JSON" 2>/dev/null; then
    return 0
  fi
  curl -sf "$BYX_REST/byx/lojas/v1/merchant?pagination.limit=200" >"$MERCHANT_QUERY_ALL_JSON"
}

extract_merchant_id_for_creator() {
  local creator=$1
  require_json_file "$MERCHANT_QUERY_ALL_JSON" "merchant-all query"
  jq -r --arg creator "$creator" '
    [
      .merchant[]?,
      .merchants[]?
    ]
    | map(select(.creator == $creator))
    | last
    | .id // empty
  ' "$MERCHANT_QUERY_ALL_JSON"
}

extract_merchant_creator_from_file() {
  local file=$1
  jq -r '.merchant.creator // .merchant.creator // empty' "$file"
}

extract_merchant_id_from_tx() {
  local txfile=$1
  require_json_file "$txfile" "merchant id extraction"
  jq -r '
    first(
      .events[]?, .tx_response.events[]?, .logs[]?.events[]?, .tx_response.logs[]?.events[]?
      | select(.type=="byx_merchant_created")
      | .attributes[]?
      | select(.key=="merchant_id")
      | .value
      | select(test("^[0-9]+$"))
    ) // empty
  ' "$txfile"
}

log_merchant_prerequisite_state() {
  local merchant_addr=$1
  log "MERCHANT_PREREQ merchant_key=$MERCHANT_KEY merchant_address=$merchant_addr loja_id=$LOJA_ID"
  if [[ -f "$MERCHANT_QUERY_BY_ID_JSON" ]] && [[ -s "$MERCHANT_QUERY_BY_ID_JSON" ]]; then
    log "MERCHANT_PREREQ query_by_id_file=$MERCHANT_QUERY_BY_ID_JSON"
  fi
  if [[ -f "$MERCHANT_QUERY_ALL_JSON" ]] && [[ -s "$MERCHANT_QUERY_ALL_JSON" ]]; then
    log "MERCHANT_PREREQ query_all_file=$MERCHANT_QUERY_ALL_JSON"
  fi
}

ensure_merchant_or_loja() {
  local merchant_addr chain_id existing_creator existing_id txhash txfile
  local -a tx_args=()
  local arg

  merchant_addr="$("$BYXD_BIN" keys show "$MERCHANT_KEY" -a --keyring-backend "$KEYRING_BACKEND")"
  chain_id="$(resolve_chain_id)"

  if query_merchant_by_id "$LOJA_ID"; then
    require_json_file "$MERCHANT_QUERY_BY_ID_JSON" "merchant query by id"
    existing_creator="$(extract_merchant_creator_from_file "$MERCHANT_QUERY_BY_ID_JSON")"
    log_merchant_prerequisite_state "$merchant_addr"
    if [[ -n "$existing_creator" && "$existing_creator" == "$merchant_addr" ]]; then
      printf '%s\n' "$LOJA_ID" >"$MERCHANT_ID_TXT"
      return 0
    fi
  fi

  query_all_merchants || die "merchant/loja prerequisite missing: failed to query merchants"
  existing_id="$(extract_merchant_id_for_creator "$merchant_addr")"
  log_merchant_prerequisite_state "$merchant_addr"
  if [[ -n "$existing_id" ]]; then
    LOJA_ID="$existing_id"
    printf '%s\n' "$LOJA_ID" >"$MERCHANT_ID_TXT"
    return 0
  fi

  while IFS= read -r arg; do
    [[ -n "$arg" ]] && tx_args+=("$arg")
  done < <(tx_common_args "$MERCHANT_KEY")
  (( ${#tx_args[@]} > 0 )) || die "failed to build merchant tx args"
  validate_real_tx_args "${tx_args[@]}"

  txhash="$(run_tx_json \
    "create-merchant" \
    "$MERCHANT_CREATE_BROADCAST_JSON" \
    "$MERCHANT_CREATE_BROADCAST_STDERR_TXT" \
    "$MERCHANT_CREATE_BROADCAST_RAW_TXT" \
    -- \
    "$BYXD_BIN" tx lojas create-merchant \
      "BYX E2E Merchant" \
      "BYX E2E Address" \
      "$merchant_addr" \
      "e2e" \
      "0000000000000000000000000000000000000000000000000000000000000000" \
      "pending" \
      "${tx_args[@]}")"

  txfile="$(mktemp)"
  wait_tx "$txhash" "$txfile"
  cp "$txfile" "$MERCHANT_CREATE_TX_JSON"
  LOJA_ID="$(extract_merchant_id_from_tx "$txfile")"
  if [[ -z "$LOJA_ID" ]]; then
    query_all_merchants || die "merchant/loja prerequisite missing: could not re-query merchants after create-merchant"
    LOJA_ID="$(extract_merchant_id_for_creator "$merchant_addr")"
  fi
  [[ -n "$LOJA_ID" ]] || die "merchant/loja prerequisite missing: could not resolve merchant id after create-merchant"
  printf '%s\n' "$LOJA_ID" >"$MERCHANT_ID_TXT"
  rm -f "$txfile"
}

json_file_is_valid() {
  local file=$1
  [[ -s "$file" ]] && jq -e . "$file" >/dev/null 2>&1
}

require_json_file() {
  local file=$1
  local context=$2
  [[ -f "$file" ]] || die "$context: missing file $file"
  [[ -s "$file" ]] || die "$context: empty file $file"
  jq -e . "$file" >/dev/null 2>&1 || die "$context: invalid json in $file"
}

is_wait_tx_retryable_output() {
  local file=$1
  [[ -f "$file" ]] || return 1
  grep -qiE 'tx not indexed|not found in cache|not found|code *= *notfound|tx.*index' "$file"
}

capture_wait_tx_response() {
  local stdout_file=$1
  local stderr_file=$2
  : >"$WAIT_TX_LAST_RESPONSE_TXT"
  if [[ -s "$stdout_file" ]]; then
    cat "$stdout_file" >>"$WAIT_TX_LAST_RESPONSE_TXT"
  fi
  if [[ -s "$stderr_file" ]]; then
    [[ -s "$WAIT_TX_LAST_RESPONSE_TXT" ]] && printf '\n' >>"$WAIT_TX_LAST_RESPONSE_TXT"
    cat "$stderr_file" >>"$WAIT_TX_LAST_RESPONSE_TXT"
  fi
}

wait_tx() {
  local txhash=$1
  local txfile=$2
  local attempt=1
  local stdout_file
  local stderr_file

  stdout_file="$(mktemp)"
  stderr_file="$(mktemp)"
  : >"$WAIT_TX_LAST_RESPONSE_TXT"

  while (( attempt <= WAIT_TX_ATTEMPTS )); do
    : >"$stdout_file"
    : >"$stderr_file"
    if "$BYXD_BIN" q tx "$txhash" --node "$BYX_RPC" -o json >"$stdout_file" 2>"$stderr_file"; then
      if json_file_is_valid "$stdout_file"; then
        mv "$stdout_file" "$txfile"
        cp "$txfile" "$CREATE_REQUEST_TX_JSON"
        rm -f "$stderr_file"
        return 0
      fi
      capture_wait_tx_response "$stdout_file" "$stderr_file"
    else
      capture_wait_tx_response "$stdout_file" "$stderr_file"
    fi

    if (( attempt < WAIT_TX_ATTEMPTS )) && is_wait_tx_retryable_output "$WAIT_TX_LAST_RESPONSE_TXT"; then
      sleep "$WAIT_TX_SLEEP_S"
      attempt=$((attempt + 1))
      continue
    fi

    if (( attempt < WAIT_TX_ATTEMPTS )); then
      sleep "$WAIT_TX_SLEEP_S"
      attempt=$((attempt + 1))
      continue
    fi
    break
  done

  rm -f "$stdout_file" "$stderr_file" "$txfile"
  die "tx query timed out waiting for indexing: txhash=$txhash node=$BYX_RPC attempts=$WAIT_TX_ATTEMPTS sleep_s=$WAIT_TX_SLEEP_S last_response_file=$WAIT_TX_LAST_RESPONSE_TXT hint=verify tx index with: curl -s $BYX_RPC/status | jq -r '.result.node_info.other.tx_index'"
}

extract_request_id_from_tx() {
  local txfile=$1
  require_json_file "$txfile" "request_id extraction"
  jq -r '
    def maybe_decoded:
      if type == "string" then ., (try @base64d catch empty) else empty end;
    def all_events:
      .events[]?, .tx_response.events[]?, .logs[]?.events[]?, .tx_response.logs[]?.events[]?;
    first(
      all_events
      | select(
          any((.type | maybe_decoded); . == "byx_payment_request_created")
          or any(.attributes[]?; any((.key | maybe_decoded); . == "request_id"))
        )
      | .attributes[]?
      | select(any((.key | maybe_decoded); . == "request_id"))
      | .value
      | maybe_decoded
      | select(test("^[0-9]+$"))
    ) // empty
  ' "$txfile"
}

log_tx_events() {
  local txfile=$1
  require_json_file "$txfile" "tx event dump"
  jq -r '
    def maybe_decoded:
      if type == "string" then ., (try @base64d catch empty) else empty end;
    [
      .events[]?,
      .tx_response.events[]?,
      .logs[]?.events[]?,
      .tx_response.logs[]?.events[]?
    ]
    | .[]
    | "event_type=" + ((.type | maybe_decoded | select(length > 0)) // "unknown")
      + " attrs="
      + (
          [
            .attributes[]? |
            (([.key | maybe_decoded | select(length > 0)] | first) // "unknown")
            + "=" +
            (([.value | maybe_decoded | select(length > 0)] | first) // "")
          ] | join(",")
        )
  ' "$txfile" >&2 || true
}

create_request() {
  local txhash txfile req_id chain_id
  local -a tx_args=()
  local arg

  while IFS= read -r arg; do
    [[ -n "$arg" ]] && tx_args+=("$arg")
  done < <(tx_common_args "$MERCHANT_KEY")
  (( ${#tx_args[@]} > 0 )) || die "failed to build tx args"
  validate_real_tx_args "${tx_args[@]}"
  chain_id="$(resolve_chain_id)"
  log_sanitized_create_request_command "$chain_id"

  txhash="$(run_tx_json \
    "create-payment-request" \
    "$CREATE_REQUEST_BROADCAST_JSON" \
    "$CREATE_REQUEST_BROADCAST_STDERR_TXT" \
    "$CREATE_REQUEST_BROADCAST_RAW_TXT" \
    -- \
    "$BYXD_BIN" tx payments create-payment-request \
      "$LOJA_ID" \
      "$AMOUNT_UBYX" \
      "${tx_args[@]}")"
  printf '%s\n' "$txhash" >"$TXHASH_CREATE_REQUEST_TXT"

  txfile="$(mktemp)"
  wait_tx "$txhash" "$txfile"
  require_json_file "$txfile" "create-payment-request tx query"

  req_id="$(extract_request_id_from_tx "$txfile" | tail -n1)"

  if [[ -z "$req_id" ]]; then
    req_id="$(curl -s "$BYX_REST/byx/payments/v1/payment_requests/by_loja/$LOJA_ID?pagination.limit=1&pagination.reverse=true" | jq -r '.payment_requests[0].id // .paymentRequests[0].id // empty')"
  fi

  if [[ -z "$req_id" ]]; then
    log "available tx events for txhash=$txhash:"
    log_tx_events "$txfile"
    rm -f "$txfile"
    die "could not resolve request_id from indexed tx events"
  fi

  printf '%s\n' "$req_id" >"$REQUEST_ID_TXT"

  jq -e '
    [
      .events[]?,
      .tx_response.events[]?,
      .logs[]?.events[]?,
      .tx_response.logs[]?.events[]?
    ]
    | .[]
    | select(.type=="byx_payment_request_created")
    | .attributes[]?
    | select(.key=="amount_ubyx")
  ' "$txfile" >/dev/null || die "create event missing amount_ubyx"
  jq -e '
    [
      .events[]?,
      .tx_response.events[]?,
      .logs[]?.events[]?,
      .tx_response.logs[]?.events[]?
    ]
    | .[]
    | .attributes[]?
    | select(.key=="amount_microbyx")
  ' "$txfile" >/dev/null && die "found deprecated amount_microbyx in create event"

  rm -f "$txfile"
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
  local -a tx_args=()
  local arg
  local txhash
  while IFS= read -r arg; do
    [[ -n "$arg" ]] && tx_args+=("$arg")
  done < <(tx_common_args "$PAYER_KEY")
  (( ${#tx_args[@]} > 0 )) || die "failed to build tx args"
  validate_real_tx_args "${tx_args[@]}"

  txhash="$(run_tx_json \
    "pay-payment-request" \
    "$PAY_REQUEST_BROADCAST_JSON" \
    "$PAY_REQUEST_BROADCAST_STDERR_TXT" \
    "$PAY_REQUEST_BROADCAST_RAW_TXT" \
    -- \
    "$BYXD_BIN" tx payments pay-payment-request \
      --request-id "$req_id" \
      "${tx_args[@]}")"
  log "PAY_REQUEST_TXHASH=$txhash"

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
  check_byxd_amount_ubyx_support
  check_stack
  check_key "$MERCHANT_KEY"
  check_key "$PAYER_KEY"
  ensure_merchant_or_loja

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
