#!/usr/bin/env bash
set -euo pipefail

RPC_URL="${RPC_URL:-http://127.0.0.1:8545}"
WS_ADDR="${WS_ADDR:-127.0.0.1:8546}"

if ! command -v curl >/dev/null 2>&1; then
  echo "curl not found" >&2
  exit 1
fi

payload_chain='{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}'
payload_net='{"jsonrpc":"2.0","method":"net_version","params":[],"id":2}'
payload_block='{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":3}'

resp="$(curl -s -X POST "$RPC_URL" -H "Content-Type: application/json" -d "$payload_chain" || true)"

if [[ -z "$resp" ]]; then
  echo "EVM_RPC_UNAVAILABLE ($RPC_URL) — habilitar EVM/JSON-RPC para eth_chainId" >&2
  exit 1
fi

chain_id="$(echo "$resp" | jq -r '.result // empty' 2>/dev/null || true)"

if [[ -n "$chain_id" && "$chain_id" != "null" ]]; then
  net_version="$(curl -s -X POST "$RPC_URL" -H "Content-Type: application/json" -d "$payload_net" | jq -r '.result // empty' 2>/dev/null || true)"
  block_number="$(curl -s -X POST "$RPC_URL" -H "Content-Type: application/json" -d "$payload_block" | jq -r '.result // empty' 2>/dev/null || true)"
  ws_status="unknown"
  if command -v nc >/dev/null 2>&1; then
    if nc -z 127.0.0.1 "${WS_ADDR##*:}" >/dev/null 2>&1; then
      ws_status="open"
    else
      ws_status="closed"
    fi
  fi
  echo "EVM_RPC_OK chain_id=$chain_id net=$net_version block=$block_number ws=$ws_status"
  exit 0
fi

echo "EVM_RPC_UNAVAILABLE ($RPC_URL) — habilitar EVM/JSON-RPC para eth_chainId" >&2
exit 1
