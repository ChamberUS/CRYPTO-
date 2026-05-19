# Sprint 11.3 — EVM Execution Plan

Objetivo desta fase: mapear passos seguros para integrar EVM (Ethermint/Evmos-style) ao BYX sem quebrar a chain.

Etapas planejadas:
1) Probe de compatibilidade (11.3): identificar versão/tag compatível da stack EVM com Cosmos SDK 0.53.3. (feito via `scripts/evm_compat_probe.sh`)
2) 11.4:
   - Fixar dependências no go.mod.
   - Wiring no app (depinject/module manager) com feature-flag EVMEnabled.
   - Store keys e module accounts EVM/feemarket.
   - Genesis/params defaults.
   - Expor JSON-RPC em 127.0.0.1:8545, validar `eth_chainId`.
3) 11.5:
   - MetaMask + contratos Solidity.
   - Scripts E2E para deploy/call.

Portas/Bindings planejadas:
- JSON-RPC: 127.0.0.1:8545 (HTTP), opcional WS 8546.
- Sem conflito com Comet RPC 26657 ou REST 1317.

Validação alvo (após integração):
- `go test ./...`
- `ignite chain serve --reset-once -v`
- `curl eth_chainId` responde.
- MetaMask conecta com chainId BYX/IAOS.
