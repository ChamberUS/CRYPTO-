#!/usr/bin/env bash
set -euo pipefail

# Inicializa perfis locais: validator, sentry, rpc-fullnode

CHAIN_ID=${CHAIN_ID:-byx-testnet-1}
DENOM=${DENOM:-ubyx}
MONIKER_VALIDATOR=${MONIKER_VALIDATOR:-byx-validator}
MONIKER_SENTRY=${MONIKER_SENTRY:-byx-sentry}
MONIKER_RPC=${MONIKER_RPC:-byx-rpc}
GENESIS_SRC=${GENESIS_SRC:-$(pwd)/genesis.json}

HOME_VAL=${HOME_VAL:-$HOME/.byx-validator}
HOME_SENTRY=${HOME_SENTRY:-$HOME/.byx-sentry}
HOME_RPC=${HOME_RPC:-$HOME/.byx-rpc}

PEX_VAL=${PEX_VAL:-0}
PEX_SENTRY=${PEX_SENTRY:-1}
PEX_RPC=${PEX_RPC:-0}
MIN_GAS_PRICES=${MIN_GAS_PRICES:-0.025ubyx}

echo "📦 Init profiles with CHAIN_ID=${CHAIN_ID} MIN_GAS_PRICES=${MIN_GAS_PRICES}"

init_node() {
  local home=$1
  local moniker=$2
  if [ -d "${home}" ]; then
    echo "ℹ️  ${home} já existe, pulando init"
    return
  fi
  byxd init "${moniker}" --chain-id "${CHAIN_ID}" --home "${home}"
  if [ -f "${GENESIS_SRC}" ]; then
    cp "${GENESIS_SRC}" "${home}/config/genesis.json"
  fi
}

init_node "${HOME_VAL}" "${MONIKER_VALIDATOR}"
init_node "${HOME_SENTRY}" "${MONIKER_SENTRY}"
init_node "${HOME_RPC}" "${MONIKER_RPC}"

VAL_ID=$(byxd tendermint show-node-id --home "${HOME_VAL}")
SENTRY_ID=$(byxd tendermint show-node-id --home "${HOME_SENTRY}")
RPC_ID=$(byxd tendermint show-node-id --home "${HOME_RPC}")

echo "validator id=${VAL_ID}"
echo "sentry id=${SENTRY_ID}"
echo "rpc id=${RPC_ID}"

configure_ports() {
  local home=$1
  local p2p_port=$2
  local rpc_port=$3
  local api_port=$4
  perl -pi -e "s#^(proxy_app = \").*#\$1tcp://127.0.0.1:26658\"#;" "${home}/config/config.toml"
  perl -pi -e "s#^(laddr = \").*26657\"#\$1tcp://127.0.0.1:${rpc_port}\"# if $.>=1 && $.<=2000;" "${home}/config/config.toml"
  perl -pi -e "s#^(laddr = \").*26656\"#\$1tcp://0.0.0.0:${p2p_port}\"# if $.>=1 && $.<=2000;" "${home}/config/config.toml"
  perl -pi -e "s#^(external_address = \").*#\$1tcp://127.0.0.1:${p2p_port}\"#;" "${home}/config/config.toml"
  perl -pi -e "s#^(max_num_inbound_peers = ).*#\${1}20#;" "${home}/config/config.toml"
  perl -pi -e "s#^(max_num_outbound_peers = ).*#\${1}40#;" "${home}/config/config.toml"
  perl -pi -e "s#^(allow_duplicate_ip = ).*#\${1}false#;" "${home}/config/config.toml"
  perl -pi -e "s#^(addr_book_strict = ).*#\${1}true#;" "${home}/config/config.toml"
  perl -pi -e "s#^(seed_mode = ).*#seed_mode = 0#;" "${home}/config/config.toml"
  perl -pi -e "s#^(pex = ).*#pex = 0#;" "${home}/config/config.toml"
  perl -pi -e "s#^(timeout_commit = \").*#\$14s\"#;" "${home}/config/config.toml"
  perl -pi -e "s#^(timeout_propose = \").*#\$13s\"#;" "${home}/config/config.toml"

  perl -pi -e "s#^(minimum-gas-prices = \").*#\$1${MIN_GAS_PRICES}\"#;" "${home}/config/app.toml"
  perl -pi -e "s#^(api.enable = ).*#${1}true#;" "${home}/config/app.toml"
  perl -pi -e "s#^(api.address = \").*#\$1tcp://127.0.0.1:${api_port}\"#;" "${home}/config/app.toml"
  perl -pi -e "s#^(grpc.enable = ).*#${1}true#;" "${home}/config/app.toml"
  perl -pi -e "s#^(grpc.address = \").*#\$1localhost:${api_port}1\"#;" "${home}/config/app.toml"
  perl -pi -e "s#^(enabled = ).*#${1}true# if $.>=1 && $.<=2000 && $ARGV eq \"${home}/config/app.toml\";" "${home}/config/app.toml"
}

configure_ports "${HOME_VAL}" 26656 26657 1317
configure_ports "${HOME_SENTRY}" 27656 27657 0
perl -pi -e "s#^(api.enable = ).*#${1}false#;" "${HOME_SENTRY}/config/app.toml"
configure_ports "${HOME_RPC}" 28656 28657 1417

# PEX per profile
perl -pi -e "s#^(pex = ).*#pex = ${PEX_VAL}#;" "${HOME_VAL}/config/config.toml"
perl -pi -e "s#^(seed_mode = ).*#seed_mode = ${PEX_VAL}#;" "${HOME_VAL}/config/config.toml"
perl -pi -e "s#^(pex = ).*#pex = ${PEX_SENTRY}#;" "${HOME_SENTRY}/config/config.toml"
perl -pi -e "s#^(seed_mode = ).*#seed_mode = ${PEX_SENTRY}#;" "${HOME_SENTRY}/config/config.toml"
perl -pi -e "s#^(pex = ).*#pex = ${PEX_RPC}#;" "${HOME_RPC}/config/config.toml"
perl -pi -e "s#^(seed_mode = ).*#seed_mode = ${PEX_RPC}#;" "${HOME_RPC}/config/config.toml"

# peers wiring
perl -pi -e "s#^(persistent_peers = \").*#\$1${SENTRY_ID}@127.0.0.1:27656\"#;" "${HOME_VAL}/config/config.toml"
perl -pi -e "s#^(seeds = \").*#\$1${VAL_ID}@127.0.0.1:26656\"#;" "${HOME_SENTRY}/config/config.toml"
perl -pi -e "s#^(persistent_peers = \").*#\$1${VAL_ID}@127.0.0.1:26656,${SENTRY_ID}@127.0.0.1:27656\"#;" "${HOME_RPC}/config/config.toml"

echo "✅ Perfis inicializados:"
echo " - validator home=${HOME_VAL}"
echo " - sentry    home=${HOME_SENTRY}"
echo " - rpc node  home=${HOME_RPC}"
