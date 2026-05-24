# Sprint 9.2 Notes (Idempotência + Relay Hardening)

## Dedupe de create-payment-request
- Fingerprint: `lojaId|amountUbyx|memo.trim()`.
- Índice KV: chave prefix + lojaID(big endian) + sha256(fingerprint) → requestID.
- Comportamento:
  - Se já houver request PENDING e não expirada com mesmo fingerprint, reusa (evento `byx_payment_request_reused`).
  - Status pago/expirado não bloqueia novo ID; índice é atualizado ao criar um novo request.
  - Ao pagar ou expirar, o índice é limpo.

## Relay: retries, DLQ, idempotência
- State (`state.json`) agora contém `sent`, `failures`, `deadLetters` (compatível com versões antigas).
- Backoff exponencial: base 500ms, cap 30s, jitter 0..250ms; limite de tentativas default 8.
- Idempotência: se `sent[requestId].eventId` for igual, não reenvia.
- Ao exceder tentativas: move para `deadLetters` e loga `DLQ` claro.
- REPLAY_DLQ=1 habilita reprocessar entradas do DLQ junto com falhas pendentes.

## Validação sugerida
- Go: `go test ./x/payments/... ./x/lojas/...`
- Dedupe script: `./scripts/test_dedupe_payment_request.sh`
- Relay (manual):
  1. Desligar mock-merchant e observar retries/backoff no log.
  2. Ligar mock-merchant; deve entregar e limpar falha.
  3. Forçar falha até atingir MAX_ATTEMPTS → aparece em `deadLetters`; set `REPLAY_DLQ=1` e reiniciar relay para reprocessar.
