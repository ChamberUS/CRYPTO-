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
- chain mode: $E2E_DIR/chain_mode.txt
- env summary: $E2E_DIR/env_summary.txt
- startup command (masked): $E2E_DIR/startup_command.txt
- failure reason: $E2E_DIR/failure_reason.txt
ART

echo "[collect] artifacts directory: $E2E_DIR"
