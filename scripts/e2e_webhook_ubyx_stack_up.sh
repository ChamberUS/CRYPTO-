#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
E2E_DIR="${E2E_DIR:-$ROOT_DIR/.e2e/webhook-ubyx}"
PIDS_FILE="$E2E_DIR/pids.env"

BYX_REST="${BYX_REST:-http://127.0.0.1:1317}"
BYX_RPC="${BYX_RPC:-http://127.0.0.1:26657}"
BYX_CHAIN_ID="${BYX_CHAIN_ID:-}"
MOCK_MERCHANT_URL="${MOCK_MERCHANT_URL:-http://127.0.0.1:4000/webhook}"
MERCHANT_WEBHOOK_SECRET="${MERCHANT_WEBHOOK_SECRET:-devsecret}"
MOCK_EVENTS_LOG_PATH="${MOCK_EVENTS_LOG_PATH:-$E2E_DIR/mock-events.jsonl}"
STATE_PATH="${STATE_PATH:-$ROOT_DIR/webhook-relay/state.json}"
STRICT_WEBHOOK="${STRICT_WEBHOOK:-1}"
LOJA_ID="${LOJA_ID:-1}"
POLL_MS="${POLL_MS:-1000}"
CHAIN_BOOT_TIMEOUT_S="${CHAIN_BOOT_TIMEOUT_S:-300}"

BYX_CHAIN_START_CMD="${BYX_CHAIN_START_CMD:-ignite chain serve --reset-once}"

mkdir -p "$E2E_DIR"
: >"$E2E_DIR/chain.log"
: >"$E2E_DIR/mock-merchant.log"
: >"$E2E_DIR/webhook-relay.log"

log() { echo "[stack-up] $*"; }

is_http_up() {
  local url=$1
  curl -sf "$url" >/dev/null 2>&1
}

port_from_url() {
  local url=$1
  echo "$url" | sed -E 's#^[a-z]+://[^:/]+:([0-9]+).*$#\1#'
}

start_chain_if_needed() {
  local rest_probe="$BYX_REST/cosmos/base/tendermint/v1beta1/syncing"
  local rpc_probe="$BYX_RPC/status"

  if is_http_up "$rest_probe" && is_http_up "$rpc_probe"; then
    log "chain already available (REST/RPC OK), skipping chain start"
    return 0
  fi

  log "starting chain with command: $BYX_CHAIN_START_CMD"
  mkdir -p "$E2E_DIR/home"
  (
    cd "$ROOT_DIR"
    nohup env HOME="$E2E_DIR/home" bash -lc "$BYX_CHAIN_START_CMD" >>"$E2E_DIR/chain.log" 2>&1 &
    echo $! >"$E2E_DIR/chain.pid"
  )

  local max_wait="$CHAIN_BOOT_TIMEOUT_S"
  local waited=0
  while (( waited < max_wait )); do
    if is_http_up "$rest_probe" && is_http_up "$rpc_probe"; then
      log "chain became healthy"
      return 0
    fi
    sleep 2
    waited=$((waited + 2))
  done

  log "chain did not become healthy in ${max_wait}s"
  log "rest url: $BYX_REST (port $(port_from_url "$BYX_REST"))"
  log "rpc url: $BYX_RPC (port $(port_from_url "$BYX_RPC"))"
  log "check log: $E2E_DIR/chain.log"
  return 1
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

start_chain_if_needed
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
