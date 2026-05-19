#!/usr/bin/env bash
set -euo pipefail

HOME_VAL=${HOME_VAL:-$HOME/.byx-validator}
HOME_SENTRY=${HOME_SENTRY:-$HOME/.byx-sentry}
HOME_RPC=${HOME_RPC:-$HOME/.byx-rpc}
LOG_DIR=${LOG_DIR:-/tmp}
MIN_GAS_PRICES=${MIN_GAS_PRICES:-0.025byx}

start_node() {
  local home=$1
  local name=$2
  local log="${LOG_DIR}/byx-${name}.log"
  if ! command -v byxd >/dev/null 2>&1; then
    echo "❌ byxd not found in PATH"
    exit 1
  fi
   if [ ! -d "${home}" ]; then
    echo "❌ home ${home} not found. Run scripts/p2p_init_profiles.sh first."
    exit 1
  fi
  echo "▶ starting ${name} (log=${log})"
  byxd start --home "${home}" --minimum-gas-prices "${MIN_GAS_PRICES}" >"${log}" 2>&1 &
  echo $! >"${LOG_DIR}/byx-${name}.pid"
}

start_node "${HOME_VAL}" "validator"
start_node "${HOME_SENTRY}" "sentry"
start_node "${HOME_RPC}" "rpc"

echo "✅ nodes started. PIDs:"
for n in validator sentry rpc; do
  if [ -f "${LOG_DIR}/byx-${n}.pid" ]; then
    echo " - ${n}: $(cat "${LOG_DIR}/byx-${n}.pid")"
  fi
done
