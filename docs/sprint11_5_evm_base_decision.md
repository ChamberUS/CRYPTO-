# Sprint 11.5 — Decisão de Base EVM (revisada)

## Decisão
**Base escolhida: Cosmos EVM (repo `github.com/cosmos/evm`) usando a chain de referência `evmd` (fork/customização).**

## Por que (motivos técnicos)
- Mantém alinhamento com a stack atual do BYX (Cosmos SDK 0.53.x / CometBFT) ao invés de forçar downgrade.
- O próprio Cosmos EVM fornece uma **reference chain (`evmd`)** exatamente para ser forkada e customizada.
- Releases do Cosmos EVM mencionam atualização para Cosmos SDK v0.53.4 e ajustes relacionados a Go v1.25+ em dependências.
- Evita o problema encontrado no probe: importar pacotes “errados/inexistentes” do cosmos/evm. Em vez disso, usamos o wiring real do `evmd`.

## Plano 11.5-A/B/C
### 11.5-A — Base EVM-first (evmd-fork)
- Trazer `evmd` para dentro do repo (como `evmd-fork/`).
- Ajustar chain-id / denom / ports / configs mantendo EVM+RPC funcionando.
- Subir node local e validar JSON-RPC/WS.

### 11.5-B — Portar módulos BYX
- Portar: x/payments, x/lojas, x/feesplit para o runtime evmd-fork.
- Ajustar params/genesis/module accounts, denom e gas.

### 11.5-C — Produto (MetaMask + deploy Solidity)
- Checklist MetaMask (rede, chainId, RPC 8545, WS 8546).
- Deploy de contrato (Hardhat ou Foundry) e tx básica.
