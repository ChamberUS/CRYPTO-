#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${ENV_FILE:-${SCRIPT_DIR}/.env}"

if [ ! -f "${ENV_FILE}" ]; then
  echo "env file not found: ${ENV_FILE}" >&2
  echo "copy ${SCRIPT_DIR}/.env.example to ${SCRIPT_DIR}/.env first" >&2
  exit 1
fi

# shellcheck source=/dev/null
source "${ENV_FILE}"

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || { echo "missing command: $1" >&2; exit 1; }
}

require_cmd "${BYXD_BIN}"
require_cmd jq

if ! id "${BYX_USER}" >/dev/null 2>&1; then
  echo "creating user ${BYX_USER}"
  useradd --system --home "${BYX_HOME}" --shell /usr/sbin/nologin "${BYX_USER}"
fi

mkdir -p "${BYX_HOME}" "${BYX_BACKUP_DIR}"
chown -R "${BYX_USER}:${BYX_GROUP}" "${BYX_HOME}" "${BYX_BACKUP_DIR}"

if [ ! -f "${BYX_HOME}/config/genesis.json" ]; then
  echo "initializing node home: ${BYX_HOME}"
  sudo -u "${BYX_USER}" "${BYXD_BIN}" init "${BYX_MONIKER}" --chain-id "${BYX_CHAIN_ID}" --home "${BYX_HOME}" >/dev/null
fi

if [ -f "${BYX_GENESIS_FILE}" ]; then
  echo "using genesis from ${BYX_GENESIS_FILE}"
  cp "${BYX_GENESIS_FILE}" "${BYX_HOME}/config/genesis.json"
fi

echo "validating genesis"
"${BYXD_BIN}" genesis validate "${BYX_HOME}/config/genesis.json"

echo "bootstrap complete"

