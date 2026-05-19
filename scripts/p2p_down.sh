#!/usr/bin/env bash
set -euo pipefail

LOG_DIR=${LOG_DIR:-/tmp}

for n in validator sentry rpc; do
  pidfile="${LOG_DIR}/byx-${n}.pid"
  if [ -f "${pidfile}" ]; then
    pid=$(cat "${pidfile}")
    echo "⛔ stopping ${n} (pid=${pid})"
    kill "${pid}" >/dev/null 2>&1 || true
    rm -f "${pidfile}"
  else
    echo "ℹ️  no pidfile for ${n}"
  fi
done

echo "✔ all nodes stopped (if they were running)"
