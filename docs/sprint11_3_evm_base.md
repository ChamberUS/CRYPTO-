# Sprint 11.3 — EVM Base (decisão e próximos passos)

## Contexto do stack atual BYX
- Cosmos SDK: v0.53.3 (go.mod)
- CometBFT: v0.38.17
- Ports padrão: RPC 26657, REST 1317, gRPC (se habilitado), monitoring Prometheus/Grafana (19090/13000).
- Módulos ativos: payments, lojas, feesplit (não mexer).

## Requisito de produto (roadmap)
- 11.3: decisão técnica + probe de compatibilidade + scaffolds de validação (sem tocar no runtime).
- 11.4: integrar EVM (Ethermint/Evmos-style) no app, com JSON-RPC 8545 respondendo `eth_chainId`.
- 11.5: MetaMask + contratos Solidity (deploy/executar) no devnet local.

## Opções avaliadas
- **Ethermint/Evmos-style** (github.com/evmos/ethermint / github.com/evmos/evmos): stack consolidada para EVM-over-Cosmos. Historically atrelada a Cosmos SDK 0.47/0.50; precisa validar compatibilidade com 0.53.
- **cosmos/evm** (github.com/cosmos/evm): WIP; também requer validação de compatibilidade.

## Plano 11.3 (executado)
- Não alteramos runtime nem go.mod.
- Criamos probe automatizado (`scripts/evm_compat_probe.sh`) para testar import/compilação das opções acima em um módulo isolado (fora do repo).
- Criamos smoke script (`scripts/evm_rpc_smoke.sh`) para testar `eth_chainId` em 8545 (esperado falhar até 11.4).
- Documentamos execução/validação.

## Plano 11.4 (próximo)
- Fixar versão/tag da stack EVM compatível com SDK 0.53 (baseado no resultado do probe).
- Adicionar dependências no go.mod.
- Wiring no app (depinject/module manager), store keys, module accounts, params defaults com feature-flag EVMEnabled.
- Expor JSON-RPC em 127.0.0.1:8545; validar `eth_chainId`.

## Plano 11.5 (posterior)
- Conectar MetaMask (RPC 8545, chainId BYX/IAOS).
- Deploy/execução de contratos Solidity.
- Testes E2E de contrato (deploy + chamada) e documentação de rede MetaMask.

## Validação alvo (quando integração estiver ativa)
- `go test ./...` passa.
- `ignite chain serve --reset-once -v` sobe sem panic.
- `byxd status`, `byxd q bank total` funcionam.
- `curl -s -X POST http://127.0.0.1:8545 -d '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' -H 'Content-Type: application/json'` responde chainId (hex).

## Observação
- Até definir uma versão compatível, evitamos mudar o go.mod. O probe documenta as opções e conflitos, servindo de base para 11.4.
