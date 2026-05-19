#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${ENV_FILE:-${SCRIPT_DIR}/.env}"

if [ -f "${ENV_FILE}" ]; then
  # shellcheck source=/dev/null
  source "${ENV_FILE}"
fi

RPC="${BYX_RPC_HTTP_URL:-http://127.0.0.1:26657}"
REST="${BYX_REST_URL:-http://127.0.0.1:1317}"

echo "== RPC status =="
curl -fsS "${RPC}/status" | jq -r '.result.node_info.network, .result.sync_info.latest_block_height'

echo "== Net info =="
curl -fsS "${RPC}/net_info" | jq -r '.result.n_peers'

echo "== REST node info =="
curl -fsS "${REST}/cosmos/base/tendermint/v1beta1/node_info" | jq -r '.default_node_info.network'

echo "healthcheck OK"

