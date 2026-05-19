#!/usr/bin/env bash
echo "⛔ Encerrando processos BYX"
pkill -f "mock-merchant" || true
pkill -f "webhook-relay" || true
pkill -f "ts-node" || true
echo "✔ Stack finalizada"
