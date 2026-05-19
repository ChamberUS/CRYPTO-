#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${ENV_FILE:-${SCRIPT_DIR}/.env}"
TEMPLATE_FILE="${SCRIPT_DIR}/systemd/byxd.service.template"

if [ ! -f "${ENV_FILE}" ]; then
  echo "env file not found: ${ENV_FILE}" >&2
  exit 1
fi

if [ ! -f "${TEMPLATE_FILE}" ]; then
  echo "template not found: ${TEMPLATE_FILE}" >&2
  exit 1
fi

# shellcheck source=/dev/null
source "${ENV_FILE}"

mkdir -p "$(dirname "${BYX_ENV_FILE}")"
cp "${ENV_FILE}" "${BYX_ENV_FILE}"
chmod 640 "${BYX_ENV_FILE}"

SERVICE_FILE="/etc/systemd/system/${BYX_SERVICE_NAME}.service"

sed \
  -e "s#{{BYX_USER}}#${BYX_USER}#g" \
  -e "s#{{BYX_GROUP}}#${BYX_GROUP}#g" \
  -e "s#{{BYX_ENV_FILE}}#${BYX_ENV_FILE}#g" \
  -e "s#{{BYXD_BIN}}#${BYXD_BIN}#g" \
  -e "s#{{BYX_HOME}}#${BYX_HOME}#g" \
  -e "s#{{BYX_MIN_GAS_PRICES}}#${BYX_MIN_GAS_PRICES}#g" \
  -e "s#{{BYX_LOG_LEVEL}}#${BYX_LOG_LEVEL}#g" \
  "${TEMPLATE_FILE}" > "${SERVICE_FILE}"

systemctl daemon-reload
systemctl enable "${BYX_SERVICE_NAME}"

echo "installed ${SERVICE_FILE}"
echo "next: systemctl start ${BYX_SERVICE_NAME}"

