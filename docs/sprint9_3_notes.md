# Sprint 9.3 Notes (Alias /iaos + Erros + Observabilidade)

## /iaos alias
- O API server reescreve prefixo `/iaos/` para `/byx/` antes do roteamento. Mesmos handlers e JSON, sem quebrar `/byx`.
- Smoke: `scripts/smoke_sprint9_1.sh` agora exige `/iaos` respondendo 200.

## Erros REST
- Módulos pagos/lojas retornam erros com códigos gRPC padrão (NotFound, InvalidArgument, FailedPrecondition) mapeados para HTTP 404/400/409 pelo gateway.
- Estrutura de erro segue o formato Cosmos base; contrato esperado: `{"error":{"code": "...", "message": "...", "details": {...}}}`.

## Eventos e trace
- Eventos de payments (`byx_payment_request_created`, `byx_payment_request_reused`, `byx_payment_request_paid`) incluem `request_id`, `loja_id`, `merchant_id`, `fingerprint_hash`, `trace_id` (hash do tx quando disponível).

## Validação rápida
- Alias: `curl -s http://127.0.0.1:1317/byx/payments/v1/payment_requests/by_loja/1?pagination.limit=1` e o mesmo em `/iaos/...` (resposta idêntica).
- Go tests: `go test ./x/payments/... ./x/lojas/...`
- E2E: `scripts/e2e_payments_webhook.sh` (poll via REST, não depende de stdout do CLI).
