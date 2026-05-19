# Sprint 11.5-A — Validação (Base EVM-first via evmd)

## Build
- `cd evmd-fork && go build ./...`

## Rodar
> Ajuste conforme Makefile/binário no evmd-fork.

- `cd evmd-fork`
- `./byxd init local --chain-id byx_1` (ou comando equivalente do evmd)
- `./byxd start`

## RPC
- `curl -s -X POST http://127.0.0.1:8545 -H 'content-type: application/json' --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' | jq`
