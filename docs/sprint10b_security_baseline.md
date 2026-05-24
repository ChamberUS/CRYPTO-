# Sprint 10B — Security Baseline (BYX / IAOS)

- **Keys & Roles**: mantenha chaves de validator offline; use sentry público e validator privado com persistent_peers.
- **Min Gas Prices**: configure `minimum-gas-prices` (ex.: `0.025ubyx`) no `app.toml` ou flag `--minimum-gas-prices`.
- **Endpoints**: bind RPC/API em `127.0.0.1` para ambientes de dev; habilite só o necessário.
- **Fee Policy**: exigir taxas mínimas e monitorar mempool.
- **Sentry/Validator**: validator sem exposição pública; sentry conectado por `persistent_peers`; opcionalmente desabilite PEX (`PEX=0`).
- **Segredos**: rotacionar `node_key`/`priv_validator_key` quando mover para produção; proteger `MERCHANT_WEBHOOK_SECRET`.
- **Telemetry**: habilitar Prometheus/telemetry local para operações; expor somente em rede confiável.
