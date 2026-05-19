# Sprint 10B — Webhook TLS (dev)

## Proxy HTTPS local
- `integrations/webhook-proxy` cria um proxy HTTPS self-signed para o mock-merchant.
- Start: `WEBHOOK_TARGET=http://127.0.0.1:3001 WEBHOOK_PROXY_PORT=3443 npm start` (ou `scripts/webhook_proxy_up.sh`).
- Use `MERCHANT_WEBHOOK_URL=https://127.0.0.1:3443/webhook` para o relay.

## Tradeoffs
- Certificado self-signed: exige `NODE_TLS_REJECT_UNAUTHORIZED=0` ou trust local durante dev.
- Produção: usar TLS válido (ACME/ingress) e proteger o endpoint com IP allowlist ou auth.
