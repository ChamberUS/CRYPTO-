# E2E – Sprint 7 (Webhook Relay com persistência)

Objetivo: provar que o relay não reenviará webhooks após reiniciar e que idempotência/persistência estão ativas.

## Pré-requisitos
- Chain BYX rodando (`ignite chain serve --reset-once`).
- Mock merchant ativo (idempotência) e relay configurado.
- CLI `byxd`, `curl`, `jq` instalados. Contas `merchant` e `payer` com saldo.

## Passo a passo
1) **Subir mock merchant** (com falhas iniciais para exercitar retry)
   ```bash
   cd webhook-relay/mock-merchant
   MERCHANT_WEBHOOK_SECRET=devsecret \
   FAIL_FIRST_N=2 \
   PORT=4000 \
   npm start
   ```

2) **Subir relay com persistência**
   ```bash
   cd webhook-relay
   REST_ENDPOINT=http://127.0.0.1:1317 \
   LOJA_ID=1 \
   MERCHANT_WEBHOOK_URL=http://127.0.0.1:4000/webhook \
   MERCHANT_WEBHOOK_SECRET=devsecret \
   STATE_PATH=./state.json \
   POLL_MS=2000 \
   npm start
   ```
   Logs esperados: config sem secret, state counts (sent/failures), polling iniciado.

3) **Criar e pagar um pedido**
   ```bash
   ./scripts/e2e_payments_webhook.sh          # modo normal, saída limpa
   DEBUG_E2E=1 ./scripts/e2e_payments_webhook.sh   # modo debug (trace + timestamps)
   STRICT_WEBHOOK=1 ./scripts/e2e_payments_webhook.sh # falha se não achar state
   STRICT_WEBHOOK=1 ./scripts/e2e_payments_webhook_ubyx.sh # smoke completo com payload ubyx + replay/hmac
   ```
   Observe o `REQUEST_ID` e aguarde `[BYX-WEBHOOK] webhook delivered...` e `[MOCK] valid webhook ...`.

4) **Testar persistência**
   - Pare o relay (Ctrl+C) após entrega do webhook.
   - Suba novamente com os mesmos envs.
   - Confirme que o pedido anterior *não* é reenviado (dedupe persistente pelo `state.json`).

5) **Retry futuro**
   - Se o mock estava falhando (`FAIL_FIRST_N`), veja logs `[BYX-WEBHOOK] queued for later retry` e depois entrega.

## Notas
- Headers enviados: `X-BYX-Signature`, `X-BYX-Idempotency-Key`, `X-BYX-Event-Id`.
- Payload inclui `event_id = <request_id>:<paid_at_unix>`.
- Estado salvo em `state.json` (configurável via `STATE_PATH`), com sent/failures e re-tentativas a cada 300s para falhas esgotadas.
- `STRICT_WEBHOOK=1` faz o script falhar se não encontrar o `event_id` no `state.json`. `DEBUG_E2E=1` habilita trace detalhado e timestamps.
