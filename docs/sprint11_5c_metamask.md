# Sprint 11.5-C — MetaMask no evmd-fork

## Rede local
- RPC HTTP: `http://127.0.0.1:8545`
- WS: `ws://127.0.0.1:8546`
- Chain ID (EVM): 262144 (`0x40000`) — confirme com `eth_chainId`
- Moeda/denom: `byx`

## Como adicionar a rede
1. Abrir MetaMask → Add network → Add a network manually.
2. Preencher:
   - Network Name: `IAOS EVM Local`
   - RPC URL: `http://127.0.0.1:8545`
   - Chain ID: `0x40000` (262144 decimal)
   - Currency Symbol: `BYX`
3. Salvar.

## Conta dev
- O script `scripts/evmd_fork_localnet_up.sh` cria a chave `alice` (keyring test) e adiciona saldo no genesis.
- Para usar no MetaMask: exporte a private key da `alice` via `evmd keys export --unarmored-hex --unsafe` (com cuidado) e importe em MetaMask.

## Checks rápidos
- `./scripts/evm_rpc_smoke.sh` → deve imprimir `EVM_RPC_OK ...` com chainId/ block.
- MetaMask deve mostrar saldo inicial (se a chave tiver sido importada).
