# Sprint 11.4 — EVM Upgrade (Cosmos EVM v0.5.x) + JSON-RPC

Meta:
- Não quebrar o runtime padrão.
- Habilitar EVM apenas via flag (EVM_ENABLED=1).
- Expor JSON-RPC em 127.0.0.1:8545 e validar com eth_chainId.

Observações:
- Cosmos EVM v0.5.0+ traz melhorias de JSON-RPC e compatibilidade com tooling (go-ethereum).
- Se surgirem erros com bytedance/sonic + Go toolchain, alinhar a versão do Go do ambiente e/ou pin/replace do sonic.
