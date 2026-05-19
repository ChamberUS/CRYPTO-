#!/usr/bin/env bash
set -euo pipefail

HOME_VAL=${HOME_VAL:-$HOME/.byx-validator}
HOME_SENTRY=${HOME_SENTRY:-$HOME/.byx-sentry}
HOME_RPC=${HOME_RPC:-$HOME/.byx-rpc}

echo "Validator RPC (127.0.0.1:26657) net_info:"
curl -sf "http://127.0.0.1:26657/net_info" | jq '{n_peers:.result.n_peers, peers:.result.peers[].node_info.id}' || echo "validator net_info unavailable"

echo "Sentry P2P (127.0.0.1:27657) net_info:"
curl -sf "http://127.0.0.1:27657/net_info" | jq '{n_peers:.result.n_peers, peers:.result.peers[].node_info.id}' || echo "sentry net_info unavailable"

echo "RPC fullnode RPC (127.0.0.1:28657) net_info:"
curl -sf "http://127.0.0.1:28657/net_info" | jq '{n_peers:.result.n_peers, peers:.result.peers[].node_info.id}' || echo "rpc net_info unavailable"

echo "API validator (127.0.0.1:1317/iaos/health) if enabled:"
curl -sf "http://127.0.0.1:1317/cosmos/base/tendermint/v1beta1/node_info" >/dev/null && echo "✅ validator API up" || echo "⚠️ validator API not reachable"

echo "API rpc-fullnode (127.0.0.1:1417) if enabled:"
curl -sf "http://127.0.0.1:1417/cosmos/base/tendermint/v1beta1/node_info" >/dev/null && echo "✅ rpc API up" || echo "⚠️ rpc API not reachable"
