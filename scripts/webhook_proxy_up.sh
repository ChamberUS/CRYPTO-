#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}/integrations/webhook-proxy"

WEBHOOK_TARGET=${WEBHOOK_TARGET:-http://127.0.0.1:3001}
WEBHOOK_PROXY_PORT=${WEBHOOK_PROXY_PORT:-3443}

echo "🔒 Starting HTTPS proxy -> ${WEBHOOK_TARGET} on port ${WEBHOOK_PROXY_PORT}"
npm install >/dev/null
WEBHOOK_TARGET="${WEBHOOK_TARGET}" WEBHOOK_PROXY_PORT="${WEBHOOK_PROXY_PORT}" npm start
