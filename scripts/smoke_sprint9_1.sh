#!/usr/bin/env bash
set -euo pipefail

REST="${REST_ENDPOINT:-http://127.0.0.1:1317}"

echo "Testing /byx and /iaos endpoints..."
echo "REST=$REST"

# Pré-check: REST vivo?
if ! curl -fsS "$REST/cosmos/base/tendermint/v1beta1/node_info" >/dev/null 2>&1; then
  echo "❌ Não consegui conectar no REST em $REST"
  echo "   - Confirme se a chain está rodando e expondo a API (porta 1317)."
  echo "   - Dica: em outro terminal, suba a chain (ex: ignite chain serve ou byxd start)."
  echo "   - Check: lsof -nP -iTCP:1317 -sTCP:LISTEN"
  exit 1
fi

echo "✅ node_info ok"

# endpoints de payments (existente hoje)
curl -fsS "$REST/byx/payments/v1/payment_requests/by_loja/1?pagination.limit=1" >/dev/null
echo "✅ /byx payments ok"

# /iaos alias deve responder igual
curl -fsS "$REST/iaos/payments/v1/payment_requests/by_loja/1?pagination.limit=1" >/dev/null
echo "✅ /iaos alias ok"

echo "SMOKE DONE"
