#!/usr/bin/env bash
set -euo pipefail

echo "Checking CometBFT RPC net_info..."
curl -sf "http://127.0.0.1:26657/net_info" | jq '.result.n_peers' && echo "✅ net_info ok" || echo "⚠️ net_info unavailable"

echo "Checking /metrics endpoints (if enabled)..."
for port in 26660 27660 28660; do
  if curl -sf "http://127.0.0.1:${port}/metrics" >/dev/null 2>&1; then
    echo "✅ metrics on ${port}"
  else
    echo "ℹ️  metrics not reachable on ${port} (enable Prometheus in config if needed)"
  fi
done

echo "Checking REST API health..."
if curl -sf "http://127.0.0.1:1317/cosmos/base/tendermint/v1beta1/node_info" >/dev/null 2>&1; then
  echo "✅ REST 1317 ok"
else
  echo "⚠️ REST 1317 not reachable"
fi
if curl -sf "http://127.0.0.1:1417/cosmos/base/tendermint/v1beta1/node_info" >/dev/null 2>&1; then
  echo "✅ REST 1417 ok"
else
  echo "ℹ️ REST 1417 not reachable (expected if rpc-fullnode off)"
fi

echo "Checking webhook-relay state.json (sent/failures/deadLetters)..."
STATE_PATH=${STATE_PATH:-./webhook-relay/state.json}
if [ -f "${STATE_PATH}" ]; then
  sent=$(jq '(.sent|length) // 0' "${STATE_PATH}")
  fail=$(jq '(.failures|length) // 0' "${STATE_PATH}")
  dlq=$(jq '(.deadLetters|length) // 0' "${STATE_PATH}")
  echo "✅ relay state: sent=${sent} failures=${fail} deadLetters=${dlq}"
else
  echo "ℹ️ state.json not found at ${STATE_PATH}"
fi

echo "Hints:"
echo "- Enable telemetry in app.toml (telemetry.enabled=true, prometheus-retain-time=60)"
echo "- Enable CometBFT prometheus in config.toml (prometheus = true, prometheus_listen_addr = \"tcp://127.0.0.1:26660\")"
