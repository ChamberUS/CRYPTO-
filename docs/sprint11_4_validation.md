# Sprint 11.4 — Validação

1) Build & tests
- `go test ./...` deve passar.

2) Subir chain com EVM habilitado (feature flag)
- Exemplo: `EVM_ENABLED=1 ignite chain serve --reset-once`
- Ou: `EVM_ENABLED=1 byxd start ...`

3) Validar JSON-RPC
- `./scripts/evm_rpc_smoke.sh` deve retornar `eth_chainId` com sucesso quando a integração estiver ativa.

4) Se falhar por dependência (sonic/Go)
- Conferir `docs/sprint11_4_evm_compat_report.md`
- Ajustar versão do Go do ambiente e/ou pin/replace de dependências.
