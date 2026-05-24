# Runbook - E2E Webhook UBYX

## Objetivo

Validar o fluxo fechado:

1. criar payment request com `amount_ubyx`
2. pagar request
3. relay detectar `PAID`
4. relay enviar webhook para mock merchant
5. mock validar idempotencia e assinatura HMAC

## Servicos necessarios

- chain BYX local (`byxd` + REST)
- mock merchant (`webhook-relay/mock-merchant`)
- webhook relay (`webhook-relay/index.ts`)

## Ordem de subida

1. Chain
2. Mock merchant
3. Relay
4. Preflight
5. Doctor (opcional)
6. E2E

## Comandos

### 1) Subir stack automatica

```bash
make stack-webhook-ubyx-up
```

Isso executa:

- tentativa de subir chain local (`ignite chain serve --reset-once` por padrao)
- mock merchant
- webhook relay

Se quiser trocar o comando da chain:

```bash
BYX_CHAIN_START_CMD="ignite chain serve --reset-once" make stack-webhook-ubyx-up
```

### 2) Preflight

```bash
make preflight-webhook-ubyx
```

### 3) Doctor

```bash
make doctor-webhook-ubyx
```

### 4) Rodar E2E completo

```bash
STRICT_WEBHOOK=1 make e2e-webhook-ubyx-full
```

### 5) Ver logs e artefatos

```bash
make stack-webhook-ubyx-logs
ls -la .e2e/webhook-ubyx
```

### 6) Derrubar stack (manual)

```bash
make stack-webhook-ubyx-down
```

## Variaveis de ambiente

- `BYX_REST` (default `http://127.0.0.1:1317`)
- `BYX_RPC` (default `http://127.0.0.1:26657`)
- `BYX_CHAIN_ID` (opcional)
- `LOJA_ID` (default `1`)
- `AMOUNT_UBYX` (default `500000`)
- `MERCHANT_KEY` (default `merchant`)
- `PAYER_KEY` (default `payer`)
- `KEYRING_BACKEND` (default `test`)
- `STATE_PATH` (default `./webhook-relay/state.json`)
- `MOCK_MERCHANT_URL` (default `http://127.0.0.1:4000/webhook`)
- `WEBHOOK_RELAY_URL` (opcional, apenas para log)
- `MOCK_EVENTS_LOG_PATH` (default `/tmp/byx_mock_events.jsonl`)
- `MERCHANT_WEBHOOK_SECRET`
- `STRICT_WEBHOOK` (default `1`)
- `BYX_CHAIN_START_CMD` (default `ignite chain serve --reset-once`)
- `CHAIN_BOOT_TIMEOUT_S` (default `300`)

## Diagnostico rapido

### REST indisponivel

- erro tipico: `REST unavailable: http://127.0.0.1:1317`
- acao:

```bash
ignite chain serve --reset-once
curl -sf http://127.0.0.1:1317/cosmos/base/tendermint/v1beta1/syncing
make doctor-webhook-ubyx
```

Se o log da chain mostrar erro de `buf.build` indisponivel:

- causa: ambiente sem acesso de rede externo para gerar proto durante `ignite chain serve`
- evidência: `.e2e/webhook-ubyx/chain.log`
- acao:
  - rodar em ambiente com internet para `buf.build`, ou
  - subir chain por comando alternativo local ja provisionado e injetar via `BYX_CHAIN_START_CMD`

### RPC indisponivel

- acao:

```bash
curl -sf http://127.0.0.1:26657/status
make doctor-webhook-ubyx
```

### Relay nao envia webhook

- confirmar relay rodando com `REST_ENDPOINT`/`MERCHANT_WEBHOOK_URL` corretos
- confirmar `STATE_PATH` acessivel
- checar `webhook-relay/state.json`

### HMAC invalido

- confirmar mesmo `MERCHANT_WEBHOOK_SECRET` no relay e no mock
- no smoke, o teste de assinatura invalida deve retornar `401` por design

### Payload antigo aparecendo

- o smoke falha se detectar `amount_microbyx`
- ajustar consumidor que ainda espera campo legado

### Idempotencia falhando

- mock deve responder `duplicate ok` no replay
- verificar headers:
  - `X-BYX-Idempotency-Key`
  - `X-BYX-Event-Id`

## Criterio de sucesso

Sucesso quando o comando termina com:

- `E2E_UBYX_OK request_id=<id>`

E os checks internos passam para:

- `amount_ubyx` presente
- `amount_microbyx` ausente no fluxo ativo
- replay aceito sem duplicar processamento
- assinatura invalida rejeitada com `401`

## Fora de escopo deste smoke

- ramp Pix/BYX
- NFT
- jogo dos pets
- dinheiro real
- testes de carga/producao

## Artefatos gerados

Após `make e2e-webhook-ubyx-full`:

- `.e2e/webhook-ubyx/chain.log`
- `.e2e/webhook-ubyx/mock-merchant.log`
- `.e2e/webhook-ubyx/webhook-relay.log`
- `.e2e/webhook-ubyx/preflight.log`
- `.e2e/webhook-ubyx/doctor.log`
- `.e2e/webhook-ubyx/e2e.log` (quando o E2E chega a executar)
- `.e2e/webhook-ubyx/state.json` (se relay iniciar)
- `.e2e/webhook-ubyx/mock-events.jsonl` (se mock receber eventos)
