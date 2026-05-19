# Sprint 11.5-C — Checklist de validação

1) Subir node local
```bash
./scripts/evmd_fork_localnet_up.sh
```
- Confirmar no log que RPC=127.0.0.1:8545 / WS=127.0.0.1:8546
- Denom usada: byx

2) Smoke do RPC
```bash
./scripts/evm_rpc_smoke.sh
```
- Deve imprimir `EVM_RPC_OK chain_id=0x40000 ...` (ou o chainId configurado no fork).

3) MetaMask
- Adicionar rede (ver `docs/sprint11_5c_metamask.md`).
- Importar a conta `alice` (exportando a private key com cuidado).
- Ver saldo > 0.

4) Deploy Solidity
```bash
cd evmd-fork/contracts/hardhat
npm install
export PRIVATE_KEY=<hex>
export EVM_RPC=http://127.0.0.1:8545
export EVM_CHAIN_ID=9000
npm run deploy
```
- Guardar endereço retornado.

5) Verificar contrato
```bash
curl -s -X POST $EVM_RPC -H 'content-type: application/json' \
  --data '{"jsonrpc":"2.0","method":"eth_getCode","params":["<endereco>", "latest"],"id":1}' | jq -r '.result'
```
- Saída diferente de `0x`.

Se todos os passos acima passarem: **Sprint 11.5-C validada**.
