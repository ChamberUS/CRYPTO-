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
GENESIS_TOML="${BYX_HOME}/config/genesis.json"

if [ ! -f "${CONFIG_TOML}" ] || [ ! -f "${APP_TOML}" ] || [ ! -f "${GENESIS_TOML}" ]; then
  echo "missing config/genesis files under ${BYX_HOME}/config" >&2
  exit 1
fi

echo "validating TOML syntax..."
python3 - "${APP_TOML}" "${CONFIG_TOML}" <<'PY'
import sys

try:
    import tomllib  # py3.11+
except ModuleNotFoundError:
    import tomli as tomllib  # type: ignore

for path in sys.argv[1:]:
    with open(path, "rb") as f:
        tomllib.load(f)
print("toml parse OK")
PY

echo "validating genesis..."
"${BYXD_BIN}" genesis validate "${GENESIS_TOML}" >/dev/null

echo "validation OK"

