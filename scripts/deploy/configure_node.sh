#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${ENV_FILE:-${SCRIPT_DIR}/.env}"

if [ ! -f "${ENV_FILE}" ]; then
  echo "env file not found: ${ENV_FILE}" >&2
  exit 1
fi

# shellcheck source=/dev/null
source "${ENV_FILE}"

CONFIG_TOML="${BYX_HOME}/config/config.toml"
APP_TOML="${BYX_HOME}/config/app.toml"

if [ ! -f "${CONFIG_TOML}" ] || [ ! -f "${APP_TOML}" ]; then
  echo "missing config files. run bootstrap first." >&2
  exit 1
fi

perl -pi -e "s#^(laddr = \").*26657\"#\$1${BYX_RPC_LADDR}\"# if $.>=1 && $.<=250;" "${CONFIG_TOML}"
perl -pi -e "s#^(laddr = \").*26656\"#\$1${BYX_P2P_LADDR}\"# if $.>=1 && $.<=250;" "${CONFIG_TOML}"
perl -pi -e "s#^(external_address = \").*#\$1${BYX_EXTERNAL_ADDRESS}\"#;" "${CONFIG_TOML}"
perl -pi -e "s#^(persistent_peers = \").*#\$1${BYX_PERSISTENT_PEERS}\"#;" "${CONFIG_TOML}"
perl -pi -e "s#^(seeds = \").*#\$1${BYX_SEEDS}\"#;" "${CONFIG_TOML}"
perl -pi -e "s#^(pex = ).*#\${1}${BYX_PEX}#;" "${CONFIG_TOML}"
perl -pi -e "s#^(prometheus = ).*#\${1}${BYX_PROMETHEUS}#;" "${CONFIG_TOML}"
perl -pi -e "s#^(prometheus_listen_addr = \").*#\$1${BYX_PROMETHEUS_LISTEN}\"#;" "${CONFIG_TOML}"
perl -pi -e "s#^(pprof_laddr = \").*#\$1${BYX_PPROF_LADDR}\"#;" "${CONFIG_TOML}"

perl -pi -e "s#^(minimum-gas-prices = \").*#\$1${BYX_MIN_GAS_PRICES}\"#;" "${APP_TOML}"
perl -pi -e "s#^(api.enable = ).*#\${1}${BYX_API_ENABLE}#;" "${APP_TOML}"
perl -pi -e "s#^(api.address = \").*#\$1${BYX_API_ADDRESS}\"#;" "${APP_TOML}"
perl -pi -e "s#^(grpc.enable = ).*#\${1}${BYX_GRPC_ENABLE}#;" "${APP_TOML}"
perl -pi -e "s#^(grpc.address = \").*#\$1${BYX_GRPC_ADDRESS}\"#;" "${APP_TOML}"
perl -pi -e "s#^(pruning = \").*#\$1${BYX_PRUNING}\"#;" "${APP_TOML}"
perl -pi -e "s#^(pruning-keep-recent = \").*#\$1${BYX_PRUNING_KEEP_RECENT}\"#;" "${APP_TOML}"
perl -pi -e "s#^(pruning-keep-every = \").*#\$1${BYX_PRUNING_KEEP_EVERY}\"#;" "${APP_TOML}"
perl -pi -e "s#^(pruning-interval = \").*#\$1${BYX_PRUNING_INTERVAL}\"#;" "${APP_TOML}"

echo "node configuration updated"
