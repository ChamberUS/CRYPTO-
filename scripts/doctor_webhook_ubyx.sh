#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
E2E_DIR="${E2E_DIR:-$ROOT_DIR/.e2e/webhook-ubyx}"

BYX_REST="${BYX_REST:-http://127.0.0.1:1317}"
BYX_RPC="${BYX_RPC:-http://127.0.0.1:26657}"
BYX_CHAIN_MODE="${BYX_CHAIN_MODE:-}"

rest_probe="$BYX_REST/cosmos/base/tendermint/v1beta1/syncing"
rpc_probe="$BYX_RPC/status"

rest_ok=0
rpc_ok=0
chain_start_attempted="unknown"

if curl -sf "$rest_probe" >/dev/null 2>&1; then rest_ok=1; fi
if curl -sf "$rpc_probe" >/dev/null 2>&1; then rpc_ok=1; fi

if [[ -z "$BYX_CHAIN_MODE" ]]; then
  if (( rest_ok == 1 && rpc_ok == 1 )); then
    BYX_CHAIN_MODE="external(auto)"
  else
    BYX_CHAIN_MODE="(unset)"
  fi
fi

if [[ -f "$E2E_DIR/env_summary.txt" ]]; then
  chain_start_attempted="$(grep '^CHAIN_START_ATTEMPTED=' "$E2E_DIR/env_summary.txt" 2>/dev/null | cut -d'=' -f2- || true)"
fi
if [[ -z "$chain_start_attempted" || "$chain_start_attempted" == "unknown" ]]; then
  case "$BYX_CHAIN_MODE" in
    external|external\(auto\)) chain_start_attempted="no" ;;
    byxd|custom|ignite) chain_start_attempted="expected yes (via stack-up)" ;;
  esac
fi

echo "=== BYX Webhook UBYX Doctor ==="
echo "BYX_CHAIN_MODE: $BYX_CHAIN_MODE"
echo "BYX_REST: $BYX_REST"
echo "BYX_RPC: $BYX_RPC"
echo "CHAIN_START_ATTEMPTED: $chain_start_attempted"

if (( rest_ok == 1 )); then
  echo "REST status: OK"
else
  echo "REST status: FAIL (porta 1317 fechada ou chain indisponivel)"
fi

if (( rpc_ok == 1 )); then
  echo "RPC status: OK"
else
  echo "RPC status: FAIL (porta 26657 fechada ou chain indisponivel)"
fi

chain_id="$(curl -sf "$rpc_probe" 2>/dev/null | jq -r '.result.node_info.network // empty' || true)"
if [[ -n "$chain_id" ]]; then
  echo "Chain-id detectado: $chain_id"
else
  echo "Chain-id detectado: (indisponivel)"
fi

echo "Portas esperadas: REST 1317, RPC 26657, mock merchant 4000"

if [[ "$BYX_CHAIN_MODE" == "ignite" ]] && [[ -f "$E2E_DIR/chain.log" ]] && grep -q "buf\\.build" "$E2E_DIR/chain.log"; then
  echo "Ignite mode failed while trying to access buf.build."
  echo "This is an environment/network/proto-cache issue."
  echo "Use BYX_CHAIN_MODE=external for an already running chain,"
  echo "BYX_CHAIN_MODE=byxd for a built binary,"
  echo "or BYX_CHAIN_MODE=custom with BYX_CHAIN_START_CMD."
fi

echo
echo "Modos disponiveis:"
echo "- external: nao sobe chain; exige REST/RPC ja ativos"
echo "- byxd: sobe com binario local (ex.: byxd start --home ...)"
echo "- custom: sobe com BYX_CHAIN_START_CMD"
echo "- ignite: sobe com ignite chain serve --reset-once (pode requerer buf.build)"
