# Sprint 10B — Stack Bootstrap (dev)

1) Ambiente
- `cp .env.example .env` (ou deixe o `scripts/dev_up.sh` gerar).
- Ajuste chaves/ports se necessário.

2) Chain
- Perfil simples: `scripts/dev_up.sh` (gera .env se faltar, sobe mock-merchant + relay).
- Perfis P2P: `scripts/p2p_init_profiles.sh` depois `scripts/p2p_up.sh`.

3) Webhook
- Mock merchant: sobe com `dev_up.sh`.
- TLS opcional: `scripts/webhook_proxy_up.sh` e set `MERCHANT_WEBHOOK_URL=https://127.0.0.1:3443/webhook`.

4) Smoke/E2E
- `scripts/smoke_sprint9_1.sh`
- `scripts/e2e_payments_webhook.sh`
- Dedupe: `scripts/test_dedupe_payment_request.sh`

5) Observabilidade
- `scripts/obs_check.sh` para checar /metrics e net_info.
