# Sprint 11.5-C — Deploy Solidity no evmd-fork

## Pré-requisitos
- Node local rodando com JSON-RPC em `http://127.0.0.1:8545` (use `scripts/evmd_fork_localnet_up.sh`).
- NodeJS + npm instalados.
- Chave com saldo e private key exportada (ex.: `alice`).

## Passo a passo (Hardhat)
```bash
cd evmd-fork/contracts/hardhat
npm install
# exporte a chave com saldo (hex) da conta genesis
export PRIVATE_KEY="<hex-da-chave>"
export EVM_RPC="http://127.0.0.1:8545"
export EVM_CHAIN_ID=9000
npm run deploy
```

Saída esperada:
```
Deploying with: 0x...
Counter deployed at: 0x...
```

## Verificar contrato
- `curl -s -X POST $EVM_RPC -H 'content-type: application/json' --data '{"jsonrpc":"2.0","method":"eth_getCode","params":["<endereco>", "latest"],"id":1}' | jq -r '.result'`
- Resultado diferente de `0x` indica que o bytecode está presente.
