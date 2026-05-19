# Sprint 9.1 — Checklist

## API / Branding
- [ ] Alias REST /iaos/* funcionando (sem quebrar /byx/*)
- [ ] Docs e exemplos usam IAOS externamente

## Idempotência
- [ ] pay-payment-request idempotente por request_id + payer
- [ ] webhook relay idempotente por event_id + request_id
- [ ] restart não duplica envios

## Observabilidade
- [ ] logs incluem request_id, loja_id, merchant_id, event_id
- [ ] erros padronizados (codespace fixo + message clara)

## Testes
- [ ] request inexistente
- [ ] merchant inexistente
- [ ] request expirada
- [ ] double pay (mesmo payer)
- [ ] retry storm webhook
