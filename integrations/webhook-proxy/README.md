# Webhook Proxy (TLS dev)

Pequeno proxy HTTPS (self-signed) para encaminhar chamadas para o mock-merchant.

## Rodar
```bash
cd integrations/webhook-proxy
npm install
WEBHOOK_TARGET=http://127.0.0.1:3001 WEBHOOK_PROXY_PORT=3443 npm start
# use MERCHANT_WEBHOOK_URL=https://127.0.0.1:3443/webhook
```
