# Sprint 9.1 Notes (B + C)

## Por que o E2E parava com "failed to parse JSON"?
O CLI pode imprimir "gas estimate: ..." antes do JSON, contaminando a captura.
Para idempotência e robustez, o E2E:
- Dispara a tx (sem parse do output)
- Confirma o estado via REST até PAID

## Por que o smoke falha?
Se o REST (1317) não estiver no ar, curl falha (normal).
