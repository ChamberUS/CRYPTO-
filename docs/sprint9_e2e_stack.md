# Sprint 9 – Stack E2E Oficial BYX / IAOS

## Objetivo
Garantir que qualquer dev consiga subir:
- blockchain
- webhook relay
- mock merchant
- fluxo E2E de pagamento

## Ordem recomendada
1. cp .env.example .env
2. Ajustar variáveis se necessário
3. scripts/dev_up.sh
4. scripts/e2e_payments_webhook.sh

## Critério de sucesso
Script E2E deve terminar com:
E2E OK: payment PAID, webhook relay should have sent notification
