#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
E2E_DIR="${E2E_DIR:-$ROOT_DIR/.e2e/webhook-ubyx}"
PIDS_FILE="$E2E_DIR/pids.env"
FAILURE_REASON_FILE="$E2E_DIR/failure_reason.txt"

BYX_REST="${BYX_REST:-http://127.0.0.1:1317}"
BYX_RPC="${BYX_RPC:-http://127.0.0.1:26657}"
BYX_CHAIN_ID="${BYX_CHAIN_ID:-}"
BYX_CHAIN_MODE="${BYX_CHAIN_MODE:-}"
BYXD_BIN="${BYXD_BIN:-byxd}"
BYX_HOME="${BYX_HOME:-}"
BYX_CHAIN_START_CMD="${BYX_CHAIN_START_CMD:-}"
MOCK_MERCHANT_URL="${MOCK_MERCHANT_URL:-http://127.0.0.1:4000/webhook}"
MERCHANT_WEBHOOK_SECRET="${MERCHANT_WEBHOOK_SECRET:-devsecret}"
MOCK_EVENTS_LOG_PATH="${MOCK_EVENTS_LOG_PATH:-$E2E_DIR/mock-events.jsonl}"
STATE_PATH="${STATE_PATH:-$ROOT_DIR/webhook-relay/state.json}"
STRICT_WEBHOOK="${STRICT_WEBHOOK:-1}"
LOJA_ID="${LOJA_ID:-1}"
POLL_MS="${POLL_MS:-1000}"
CHAIN_BOOT_TIMEOUT_S="${CHAIN_BOOT_TIMEOUT_S:-300}"

CHAIN_START_ATTEMPTED=0
CHAIN_STARTUP_COMMAND=""

mkdir -p "$E2E_DIR"
: >"$E2E_DIR/chain.log"
: >"$E2E_DIR/mock-merchant.log"
: >"$E2E_DIR/webhook-relay.log"
rm -f "$FAILURE_REASON_FILE"

log() { echo "[stack-up] $*"; }
fail() {
  local msg=$1
  echo "$msg" >"$FAILURE_REASON_FILE"
  write_runtime_artifacts >/dev/null 2>&1 || true
  log "ERROR: $msg"
  exit 1
}

is_http_up() {
  local url=$1
  curl -sf "$url" >/dev/null 2>&1
}

port_from_url() {
  local url=$1
  echo "$url" | sed -E 's#^[a-z]+://[^:/]+:([0-9]+).*$#\1#'
}

sanitize_endpoint() {
  local url=$1
  if [[ "$url" =~ ^https?://(127\.0\.0\.1|localhost)(:[0-9]+)?(/.*)?$ ]]; then
    echo "$url"
    return
  fi
  echo "$url" | sed -E 's#^(https?://)[^/:]+#\1<redacted-host>#'
}

mask_sensitive() {
  local text=$1
  echo "$text" | sed -E \
    -e 's#([A-Za-z_]*(SECRET|TOKEN|PASSWORD|KEY)[A-Za-z_]*=)[^ ]+#\1***#gI' \
    -e 's#(https?://)[^/@]+@#\1***@#g'
}

write_runtime_artifacts() {
  cat >"$E2E_DIR/chain_mode.txt" <<EOF
$BYX_CHAIN_MODE
EOF

  cat >"$E2E_DIR/env_summary.txt" <<EOF
BYX_CHAIN_MODE=$BYX_CHAIN_MODE
BYX_REST=$(sanitize_endpoint "$BYX_REST")
BYX_RPC=$(sanitize_endpoint "$BYX_RPC")
BYX_CHAIN_ID=${BYX_CHAIN_ID:-<unset>}
BYXD_BIN=${BYXD_BIN:-<unset>}
BYX_HOME=${BYX_HOME:-<unset>}
CHAIN_START_ATTEMPTED=$CHAIN_START_ATTEMPTED
EOF

  cat >"$E2E_DIR/startup_command.txt" <<EOF
$(mask_sensitive "${CHAIN_STARTUP_COMMAND:-<not-started>}")
EOF
}

is_chain_healthy() {
  local rest_probe="$BYX_REST/cosmos/base/tendermint/v1beta1/syncing"
  local rpc_probe="$BYX_RPC/status"
  is_http_up "$rest_probe" && is_http_up "$rpc_probe"
}

detect_chain_mode() {
  if [[ -n "$BYX_CHAIN_MODE" ]]; then
    case "$BYX_CHAIN_MODE" in
      external|byxd|custom|ignite) return 0 ;;
      *)
        fail "invalid BYX_CHAIN_MODE='$BYX_CHAIN_MODE' (use: external|byxd|custom|ignite)"
        ;;
    esac
  fi

  if is_chain_healthy; then
    BYX_CHAIN_MODE="external"
    log "BYX_CHAIN_MODE not set; auto-selected 'external' because REST/RPC are healthy"
    return 0
  fi

  fail "BYX_CHAIN_MODE is not set and REST/RPC are unavailable. Set BYX_CHAIN_MODE=external|byxd|custom|ignite."
}

write_ignite_buf_hint() {
  log "Ignite mode failed while trying to access buf.build."
  log "This is an environment/network/proto-cache issue."
  log "Use BYX_CHAIN_MODE=external for an already running chain,"
  log "BYX_CHAIN_MODE=byxd for a built binary,"
  log "or BYX_CHAIN_MODE=custom with BYX_CHAIN_START_CMD."
}

start_chain_with_command() {
  local command=$1
  CHAIN_START_ATTEMPTED=1
  CHAIN_STARTUP_COMMAND="$command"
  log "starting chain in mode '$BYX_CHAIN_MODE'"
  log "startup command: $(mask_sensitive "$CHAIN_STARTUP_COMMAND")"
  mkdir -p "$E2E_DIR/home"
  (
    cd "$ROOT_DIR"
    nohup env HOME="$E2E_DIR/home" bash -lc "$command" >>"$E2E_DIR/chain.log" 2>&1 &
    echo $! >"$E2E_DIR/chain.pid"
  )
}

wait_chain_healthy() {
  local chain_pid
  chain_pid="$(cat "$E2E_DIR/chain.pid" 2>/dev/null || true)"

  local max_wait="$CHAIN_BOOT_TIMEOUT_S"
  local waited=0
  while (( waited < max_wait )); do
    if is_chain_healthy; then
      log "chain became healthy"
      return 0
    fi
    if [[ -n "$chain_pid" ]] && ! kill -0 "$chain_pid" >/dev/null 2>&1; then
      break
    fi
    sleep 2
    waited=$((waited + 2))
  done

  log "chain did not become healthy in ${max_wait}s (mode=$BYX_CHAIN_MODE)"
  log "rest url: $BYX_REST (port $(port_from_url "$BYX_REST"))"
  log "rpc url: $BYX_RPC (port $(port_from_url "$BYX_RPC"))"
  log "check log: $E2E_DIR/chain.log"
  if [[ "$BYX_CHAIN_MODE" == "ignite" ]] && grep -q "buf\\.build" "$E2E_DIR/chain.log"; then
    write_ignite_buf_hint
  fi
  return 1
}

build_byxd_start_command() {
  local parts=("$BYXD_BIN" "start")
  if [[ -n "$BYX_HOME" ]]; then
    parts+=("--home" "$BYX_HOME")
  fi
  printf '%q ' "${parts[@]}"
}

ensure_external_chain() {
  if ! is_chain_healthy; then
    fail "external mode requires an already running chain (REST/RPC unavailable at $BYX_REST and $BYX_RPC)."
  fi
  log "external mode: using existing chain (REST/RPC OK)"
}

start_chain_by_mode() {
  if [[ "$BYX_CHAIN_MODE" == "external" ]]; then
    ensure_external_chain
    return 0
  fi

  if is_chain_healthy; then
    log "chain already healthy; skipping startup in mode '$BYX_CHAIN_MODE'"
    return 0
  fi

  case "$BYX_CHAIN_MODE" in
    byxd)
      local byxd_cmd
      byxd_cmd="$(build_byxd_start_command)"
      start_chain_with_command "$byxd_cmd"
      wait_chain_healthy || fail "byxd mode failed to make REST/RPC healthy"
      ;;
    custom)
      if [[ -z "$BYX_CHAIN_START_CMD" ]]; then
        fail "custom mode requires BYX_CHAIN_START_CMD"
      fi
      start_chain_with_command "$BYX_CHAIN_START_CMD"
      wait_chain_healthy || fail "custom mode failed to make REST/RPC healthy"
      ;;
    ignite)
      local ignite_cmd="${BYX_CHAIN_START_CMD:-ignite chain serve --reset-once}"
      start_chain_with_command "$ignite_cmd"
      wait_chain_healthy || fail "ignite mode failed to make REST/RPC healthy"
      ;;
    *)
      fail "invalid BYX_CHAIN_MODE='$BYX_CHAIN_MODE'"
      ;;
  esac
}

start_mock_merchant() {
  if is_http_up "$MOCK_MERCHANT_URL"; then
    log "mock merchant already available at $MOCK_MERCHANT_URL"
    return 0
  fi

  log "starting mock merchant"
  (
    cd "$ROOT_DIR/webhook-relay/mock-merchant"
    nohup env \
      PORT="$(port_from_url "$MOCK_MERCHANT_URL")" \
      MERCHANT_WEBHOOK_SECRET="$MERCHANT_WEBHOOK_SECRET" \
      MOCK_EVENTS_LOG_PATH="$MOCK_EVENTS_LOG_PATH" \
      npm start >>"$E2E_DIR/mock-merchant.log" 2>&1 &
    echo $! >"$E2E_DIR/mock-merchant.pid"
  )

  local max_wait=20
  local waited=0
  while (( waited < max_wait )); do
    if is_http_up "$MOCK_MERCHANT_URL"; then
      log "mock merchant healthy"
      return 0
    fi
    sleep 1
    waited=$((waited + 1))
  done

  log "mock merchant did not become healthy"
  log "check log: $E2E_DIR/mock-merchant.log"
  return 1
}

start_webhook_relay() {
  if [[ -f "$STATE_PATH" ]] && pgrep -f "node --loader ts-node/esm index.ts" >/dev/null 2>&1; then
    log "webhook relay appears already running, reusing"
    return 0
  fi

  log "starting webhook relay"
  (
    cd "$ROOT_DIR/webhook-relay"
    nohup env \
      REST_ENDPOINT="$BYX_REST" \
      LOJA_ID="$LOJA_ID" \
      MERCHANT_WEBHOOK_URL="$MOCK_MERCHANT_URL" \
      MERCHANT_WEBHOOK_SECRET="$MERCHANT_WEBHOOK_SECRET" \
      STATE_PATH="$STATE_PATH" \
      MOCK_EVENTS_LOG_PATH="$MOCK_EVENTS_LOG_PATH" \
      STRICT_WEBHOOK="$STRICT_WEBHOOK" \
      POLL_MS="$POLL_MS" \
      npm start >>"$E2E_DIR/webhook-relay.log" 2>&1 &
    echo $! >"$E2E_DIR/webhook-relay.pid"
  )

  local max_wait=20
  local waited=0
  while (( waited < max_wait )); do
    if [[ -f "$STATE_PATH" ]]; then
      log "webhook relay bootstrapped state file"
      return 0
    fi
    sleep 1
    waited=$((waited + 1))
  done

  log "webhook relay did not bootstrap state file"
  log "state path: $STATE_PATH"
  log "check log: $E2E_DIR/webhook-relay.log"
  return 1
}

detect_chain_mode
start_chain_by_mode
write_runtime_artifacts
start_mock_merchant
start_webhook_relay

{
  echo "E2E_DIR=$E2E_DIR"
  echo "STATE_PATH=$STATE_PATH"
  echo "MOCK_EVENTS_LOG_PATH=$MOCK_EVENTS_LOG_PATH"
  [[ -f "$E2E_DIR/chain.pid" ]] && echo "CHAIN_PID=$(cat "$E2E_DIR/chain.pid")"
  [[ -f "$E2E_DIR/mock-merchant.pid" ]] && echo "MOCK_MERCHANT_PID=$(cat "$E2E_DIR/mock-merchant.pid")"
  [[ -f "$E2E_DIR/webhook-relay.pid" ]] && echo "WEBHOOK_RELAY_PID=$(cat "$E2E_DIR/webhook-relay.pid")"
} >"$PIDS_FILE"

log "stack is up"
log "logs dir: $E2E_DIR"
