#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
MON_DIR="$ROOT_DIR/infra/monitoring"

if ! command -v docker >/dev/null 2>&1; then
  echo "docker not found in PATH" >&2
  exit 1
fi

if ! command -v docker compose >/dev/null 2>&1; then
  echo "docker compose not found (Docker CLI v2 required)" >&2
  exit 1
fi

cd "$MON_DIR"
docker compose up -d

PROM_URL="http://localhost:19090/-/ready"
GRAF_URL="http://localhost:13000/api/health"

if curl -fsS "$PROM_URL" >/dev/null 2>&1; then
  echo "PROMETHEUS_OK http://localhost:19090"
else
  echo "PROMETHEUS_NOT_READY ($PROM_URL)" >&2
  exit 1
fi

if curl -fsS "$GRAF_URL" >/dev/null 2>&1; then
  echo "GRAFANA_OK http://localhost:13000"
else
  echo "GRAFANA_NOT_READY ($GRAF_URL)" >&2
  exit 1
fi
