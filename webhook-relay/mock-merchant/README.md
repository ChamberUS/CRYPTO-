# Mock Merchant Webhook

Servidor HTTP mínimo para validar o fluxo do relay. Sem dependências externas.

## Rodando
```bash
cd webhook-relay/mock-merchant
MERCHANT_WEBHOOK_SECRET=devsecret \
FAIL_FIRST_N=2 \
PORT=4000 \
npm start
```

Variáveis:
- `MERCHANT_WEBHOOK_SECRET` (ou `SECRET`): segredo HMAC compartilhado com o relay.
- `FAIL_FIRST_N`: número de requisições iniciais que retornam 500 para testar retry/backoff.
- `PORT`: porta HTTP (default 4000).

Logs esperados:
- `[MOCK] failing intentionally (n/N)` nas primeiras chamadas quando `FAIL_FIRST_N` > 0.
- `[MOCK] invalid signature` quando o HMAC não bate.
- `[MOCK] valid webhook request_id=... amount=...` quando a assinatura é válida.
