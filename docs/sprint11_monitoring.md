# Sprint 11.0–11.2 — Monitoramento local (Prometheus + Grafana)

## Stack
- Prometheus: http://localhost:19090
- Grafana: http://localhost:13000
- Compose files: `infra/monitoring/`

## Como subir
1. Subir a chain (exemplo):
   ```bash
   ignite chain serve --reset-once
   ```
2. Subir Prometheus + Grafana:
   ```bash
   scripts/monitoring_up.sh
   ```
   Saída esperada:
   - `PROMETHEUS_OK http://localhost:19090`
   - `GRAFANA_OK http://localhost:13000`

## Ajustar targets de scrape
- Por padrão, os alvos usam:
  - `COMET_METRICS_TARGET=host.docker.internal:26660`
  - `APP_METRICS_TARGET=host.docker.internal:26660`
- Para mudar, exporte antes de subir:
  ```bash
  export COMET_METRICS_TARGET=host.docker.internal:26660
  export APP_METRICS_TARGET=host.docker.internal:26660
  scripts/monitoring_up.sh
  ```
  (Valores dependem de onde o CometBFT/telemetry expõem `/metrics`.)

## Validar rapidamente
- Acesse Prometheus: http://localhost:19090 e busque a série `up`.
- Acesse Grafana: http://localhost:13000 (login default grafana/grafana se solicitado).

## Telemetry do app (próximo passo)
- Se o `byxd` expuser `/metrics` (telemetry Prometheus do Cosmos SDK), ajuste `APP_METRICS_TARGET` para apontar para a porta correta.
- Caso não esteja habilitado ainda, habilitar telemetry no `app.toml` ou equivalente (Cosmos SDK v0.53 usa seção `telemetry`).
