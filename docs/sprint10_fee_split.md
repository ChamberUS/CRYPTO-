# Sprint 10 — Fee Split (60/30/10) sobre gas fees

## O que faz
- Aplica split **somente** sobre as gas fees acumuladas no `FeeCollector` antes de `x/distribution`.
- Percentual padrão (bps): 6000 validators, 3000 treasury, 1000 burn (total 10000).
- Allowlist de denoms: default `["byx"]`; allowlist vazia aplica a todas as denoms.
- Executa no `BeginBlock` do módulo `feesplit` (ordenado antes de `distribution`).
- Tesouro: module account `treasury` (sem perms especiais). Burn: via module account `feesplit` com permissão `burner`.

## Parâmetros
- `enabled` (bool, default true)
- `split_bps_*` (validators/treasury/burn) devem somar 10000.
- `denoms_allowlist` (se vazia, aplica a todos; não aceitar strings vazias).

## Validação rápida
1) Tests Go:
```bash
GOCACHE=$PWD/.gocache go test ./x/feesplit/... ./x/payments/... ./x/lojas/...
```
2) Smoke do split on-chain (node rodando):
```bash
scripts/test_fee_split.sh
# Esperado: FEE_SPLIT_OK
```
O script envia um tx com fee `10000byx`, espera 1 bloco e verifica:
- Tesouro recebeu +3000 byx
- Supply caiu 1000 byx (burn)
3) E2E pagamentos/webhook deve continuar intacto.

## Notas de determinismo
- Sem uso de rand/tempo externo/REST.
- Cálculo por denom: `treasury=floor(total*3000/10000)`, `burn=floor(total*1000/10000)`, resto fica no `FeeCollector` para distribuição.
