# BYX Webhook Relay

Pequeno worker em Node/TS que faz polling no endpoint REST do modulo `payments` e dispara webhooks para o lojista quando um pedido muda para `PAID`.

## Variaveis de ambiente

- `REST_ENDPOINT` - endpoint REST do no BYX (ex.: `http://localhost:1317`).
- `MERCHANT_WEBHOOK_URL` - URL HTTP do backend do lojista para receber POST.
- `MERCHANT_WEBHOOK_SECRET` - segredo compartilhado para assinar o payload.
- `LOJA_ID` - ID numerico da loja monitorada.
- `STATE_PATH` - arquivo de estado local para idempotencia e retry.

## Rodando

Use Node 18+ (fetch nativo). Exemplo com `ts-node`:

```bash
cd webhook-relay
REST_ENDPOINT=http://localhost:1317 \
MERCHANT_WEBHOOK_URL=http://localhost:4000/webhook \
MERCHANT_WEBHOOK_SECRET=supersegredo \
LOJA_ID=1 \
npm start
```

## Payload enviado

```json
{
  "request_id": 12,
  "loja_id": 1,
  "amount_ubyx": 2500000,
  "payer": "byx1...",
  "paid_at_unix": 1736533323,
  "event_id": "12:1736533323",
  "trace_id": "12:1736533323"
}
```

Headers:

- `X-BYX-Signature`
- `X-BYX-Idempotency-Key`
- `X-BYX-Event-Id`

Assinatura HMAC-SHA256: `hex(hmac_sha256(secret, body))`.

## Exemplo de verificacao

```js
import { createHmac } from "crypto";

function isValidSignature(body, signature, secret) {
  const expected = createHmac("sha256", secret).update(body).digest("hex");
  return expected === signature;
}
```
