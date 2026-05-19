# Sprint 11.3 — Validação esperada

Com a integração EVM pendente, as validações desta fase focam em scaffolds e compat:

- `GOCACHE=$PWD/.gocache go test ./...` deve passar.
- `./scripts/monitoring_up.sh` deve subir Prometheus (19090) e Grafana (13000).
- `./scripts/evm_compat_probe.sh` deve gerar `docs/sprint11_3_evm_compat_report.md` com o status das opções testadas.
- `./scripts/evm_rpc_smoke.sh`:
  - Se JSON-RPC ainda não estiver habilitado, deve falhar com mensagem clara (“esperado falhar na 11.3; 11.4 habilita JSON-RPC”).
  - Depois da integração (11.4+), deve imprimir `EVM_RPC_OK`.
