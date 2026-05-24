#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
E2E_DIR="${E2E_DIR:-$ROOT_DIR/.e2e/webhook-ubyx}"
PIDS_FILE="$E2E_DIR/pids.env"

log() { echo "[stack-down] $*"; }

kill_pid_file() {
  local file=$1
  local label=$2
  if [[ ! -f "$file" ]]; then
    return 0
  fi
  local pid
  pid="$(cat "$file" 2>/dev/null || true)"
  if [[ -n "$pid" ]] && kill -0 "$pid" >/dev/null 2>&1; then
    kill "$pid" >/dev/null 2>&1 || true
    sleep 1
    kill -9 "$pid" >/dev/null 2>&1 || true
    log "stopped $label pid=$pid"
  fi
  rm -f "$file"
}

kill_pid_file "$E2E_DIR/webhook-relay.pid" "webhook relay"
kill_pid_file "$E2E_DIR/mock-merchant.pid" "mock merchant"
kill_pid_file "$E2E_DIR/chain.pid" "chain"

rm -f "$PIDS_FILE"
log "stack down complete"
