# BYX Webhook Relay

Pequeno worker em Node/TS que faz polling no endpoint REST do módulo `payments` e dispara webhooks para o lojista quando um pedido muda para `PAID`.

## Variáveis de ambiente

- `REST_ENDPOINT` – endpoint REST do nó BYX (ex.: `http://localhost:1317`).
- `MERCHANT_WEBHOOK_URL` – URL HTTP do seu backend que receberá o POST.
- `MERCHANT_WEBHOOK_SECRET` – segredo compartilhado para assinar o payload.
- `LOJA_ID` – ID numérico da loja que terá os pedidos monitorados.

## Rodando

Use Node 18+ (fetch nativo). Exemplo com `ts-node`:

```bash
cd webhook-relay
REST_ENDPOINT=http://localhost:1317 \
MERCHANT_WEBHOOK_URL=http://localhost:4000/webhook \
MERCHANT_WEBHOOK_SECRET=supersegredo \
LOJA_ID=1 \
npx ts-node --transpile-only index.ts
```

Se preferir transpilar, basta rodar `tsc` apontando para `index.ts` ou copiar o código para um arquivo `.js` (o código é compatível).

O worker faz polling a cada 2s e envia um POST com:

```json
{
  "request_id": 12,
  "loja_id": 1,
  "amount": 2500000,
  "payer": "byx1...",
  "paid_at": 1736533323
}
```

Assinatura HMAC-SHA256: header `X-BYX-Signature` contendo `hex(hmac_sha256(secret, body))`.

## Exemplo de verificação

```js
import { createHmac } from "crypto";

function isValidSignature(body, signature, secret) {
  const expected = createHmac("sha256", secret).update(body).digest("hex");
  return expected === signature;
}
```

Em Express:

```js
app.post("/webhook", express.json({ type: "*/*" }), (req, res) => {
  const sig = req.header("X-BYX-Signature") || "";
  const raw = JSON.stringify(req.body);
  if (!isValidSignature(raw, sig, process.env.MERCHANT_WEBHOOK_SECRET)) {
    return res.status(401).send("invalid signature");
  }
  // processa req.body
  res.sendStatus(200);
});
```
