# E2E – Sprint 6.2 (Payment Request → Pay → Webhook)

Fluxo completo usando chain local, mock merchant e relay com dedupe/retry.

## Pré-requisitos
- `ignite chain serve` ativo na raiz (reset inicial recomendado: `ignite chain serve --reset-once`).
- CLI `byxd`, `curl`, `jq` instalados.
- Contas no keyring para o lojista (`MERCHANT_KEY`, default `merchant`) e pagador (`PAYER_KEY`, default `payer`) com saldo suficiente.

## Passo a passo (≈1 min)
1) **Subir chain**
   ```bash
   ignite chain serve --reset-once
   ```

2) **Mock merchant**
   ```bash
   cd webhook-relay/mock-merchant
   MERCHANT_WEBHOOK_SECRET=devsecret \
   FAIL_FIRST_N=2 \
   PORT=4000 \
   npm start
   ```

3) **Webhook relay**
   ```bash
   cd webhook-relay
   REST_ENDPOINT=http://127.0.0.1:1317 \
   LOJA_ID=1 \
   MERCHANT_WEBHOOK_URL=http://127.0.0.1:4000/webhook \
   MERCHANT_WEBHOOK_SECRET=devsecret \
   POLL_MS=2000 \
   npm start
   ```
   Logs esperados: `config REST=... LOJA_ID=...` e `[BYX-WEBHOOK] polling iniciado...`.

4) **Rodar script E2E (legado)**
   ```bash
   ./scripts/e2e_payments_webhook.sh
   ```
   - Cria PaymentRequest, paga e aguarda `PAID` (timeout 15s).

5) **Rodar smoke E2E ubyx (recomendado)**
   ```bash
   STRICT_WEBHOOK=1 \
   STATE_PATH=./webhook-relay/state.json \
   MOCK_EVENTS_LOG_PATH=/tmp/byx_mock_events.jsonl \
   MERCHANT_WEBHOOK_SECRET=devsecret \
   ./scripts/e2e_payments_webhook_ubyx.sh
   ```
   - Valida `amount_ubyx` no create/query/qr/webhook.
   - Falha se encontrar `amount_microbyx` no fluxo ativo.
   - Testa replay (idempotência no mock) e assinatura inválida (HTTP 401).

6) **Verificar logs**
   - Relay: `[BYX-WEBHOOK] request <id> detected as PAID`, retries se o mock estiver falhando.
   - Mock: `[MOCK] failing intentionally (n/N)` (primeiros chamados), depois `[MOCK] valid webhook request_id=... amount=...`.
   - Dedupe: apenas um `[MOCK] valid webhook` por `request_id`.

## Variáveis úteis
- `REST`: endpoint REST (default `http://127.0.0.1:1317`).
- `LOJA_ID`: loja alvo (default `1`).
- `MERCHANT_KEY` / `PAYER_KEY`: nomes no keyring `byxd keys list`.
- `AMOUNT_UBYX`: valor usado pelo script (default `500000` ubyx = 0.5 BYX).
- `FAIL_FIRST_N`: número de falhas iniciais no mock para exercitar retry/backoff.

## Observações
- O webhook-relay já possui dedupe em memória: não reenviará webhooks `PAID` para o mesmo `request_id`.
- Retry/backoff: 5 tentativas com 500ms, 1s, 2s, 4s, 8s e timeout de 3s por chamada.
