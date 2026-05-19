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

STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
OUT_DIR="${BYX_BACKUP_DIR}/${STAMP}"
mkdir -p "${OUT_DIR}"

echo "creating backup at ${OUT_DIR}"

tar -czf "${OUT_DIR}/config.tar.gz" -C "${BYX_HOME}" config
tar -czf "${OUT_DIR}/data.tar.gz" -C "${BYX_HOME}" data

sha256sum "${OUT_DIR}/config.tar.gz" > "${OUT_DIR}/config.tar.gz.sha256"
sha256sum "${OUT_DIR}/data.tar.gz" > "${OUT_DIR}/data.tar.gz.sha256"

find "${BYX_BACKUP_DIR}" -mindepth 1 -maxdepth 1 -type d -mtime +"${BYX_BACKUP_KEEP_DAYS}" -exec rm -rf {} \;

echo "backup complete"

