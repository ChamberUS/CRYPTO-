# Sprint 11.5-B — Port de módulos BYX para iaos-evmd

## O que foi portado
- `x/feesplit` (commit anterior) — integrado ao app com Begin/End/Genesis.
- `x/lojas` (este passo) — keeper, types, params, genesis, CLI/autocli e simulações.
- `testutil/sample` — utilitários usados pelas simulações de `x/lojas`.
- `x/payments` (este passo) — keeper, types, genesis, params, CLI/autocli e dedupe logic.

## Integração no runtime `evmd-fork`
- `app.go`: adicionados store key, keeper e módulo `lojas` no ModuleManager, orders de Begin/End/InitGenesis/ExportGenesis.
- `app.go`: adicionados store key, keeper e módulo `payments` (dependência direta de `lojas` e `bank`), nas ordens de Begin/End/InitGenesis/ExportGenesis.
- Keeper instancia com `AccountKeeper`/`BankKeeper` e authority `gov` (mesmo padrão do módulo original).
- Imports e paths ajustados para `github.com/buynnex/iaos-evmd/...`.

## Como validar
```bash
cd evmd-fork
GOTOOLCHAIN=local GOWORK=off GOPROXY="https://proxy.golang.org,direct" GOSUMDB=off go mod tidy
GOTOOLCHAIN=local GOWORK=off go build ./...
```
Logs gravados em:
- `docs/evmd_fork_tidy_11_5b_lojas.log`
- `docs/evmd_fork_build_11_5b_lojas.log`
- `docs/evmd_fork_tidy_11_5b_payments.log`
- `docs/evmd_fork_build_11_5b_payments.log`

## Próximos passos
- Validar fluxo de genesis/queries dos módulos portados no fork (CLI ou testes unitários em etapa posterior).
