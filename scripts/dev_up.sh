#!/usr/bin/env bash
set -e

echo "🔼 Subindo stack BYX (Sprint 9)"

# Carregar env
if [ ! -f .env ]; then
  if [ -f .env.example ]; then
    cp .env.example .env
    echo "ℹ️  .env não encontrado. Gerado automaticamente a partir de .env.example"
  else
    echo "⚠️  .env e .env.example não encontrados. Usando variáveis padrão."
  fi
fi

if [ -f .env ]; then
  # shellcheck disable=SC2046
  export $(grep -v '^#' .env | xargs)
fi

# Mock Merchant
echo "▶ Mock Merchant"
(cd integrations/mock-merchant && npm install && npm run dev) &

# Webhook Relay
echo "▶ Webhook Relay"
(cd webhook-relay && npm run start) &

echo "✔ Stack subida"
echo "Abra outro terminal para rodar o E2E"
wait
