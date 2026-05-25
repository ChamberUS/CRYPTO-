# Runbook - E2E Webhook UBYX

## Objetivo

Validar o fluxo fechado:

1. criar payment request com `amount_ubyx`
2. pagar request
3. relay detectar `PAID`
4. relay enviar webhook para mock merchant
5. mock validar idempotencia e assinatura HMAC

## Modos de startup da chain (`BYX_CHAIN_MODE`)

- `external`: nao sobe chain; exige REST/RPC ja ativos.
- `byxd`: sobe chain com binario local (`BYXD_BIN`, default `byxd`).
- `custom`: sobe chain com comando custom (`BYX_CHAIN_START_CMD`).
- `ignite`: sobe chain com `ignite chain serve --reset-once` (pode exigir acesso a `buf.build`).

Quando `BYX_CHAIN_MODE` nao for informado:

- se REST/RPC estiverem ativos, o `stack-up` usa `external` automaticamente;
- se REST/RPC nao estiverem ativos, o script falha com orientacao para escolher `external|byxd|custom|ignite`.

## Servicos necessarios

- chain BYX (local ou remota) com REST/RPC acessiveis
- mock merchant (`webhook-relay/mock-merchant`)
- webhook relay (`webhook-relay/index.ts`)

## Nota de bootstrap do relay (Node 20+)

Causa raiz observada em falha recente:

- startup antigo do relay: `node --loader ts-node/esm index.ts`
- em Node 20, esse loader ficou instavel no ambiente e o processo abortava antes do bootstrap;
- `state.json` nao era criado e o stack-up falhava em `webhook relay did not bootstrap state file`.

Correcao aplicada:

- startup novo do relay: `tsx index.ts`
- `STATE_PATH` padrao do fluxo E2E: `.e2e/webhook-ubyx/state.json`
- relay agora cria automaticamente o diretorio pai de `STATE_PATH` antes de gravar estado.

## Nota de saude do mock merchant

Causa raiz observada em falha recente de stack-up:

- o mock merchant respondia apenas `POST /webhook`;
- o health check anterior testava rota sem endpoint de saude e recebia `404`;
- isso gerava `mock merchant did not become healthy`.

Agora o mock expoe:

- `GET /health` -> `200` com `{"ok":true,"service":"mock-merchant"}`
- `GET /healthz` -> alias do mesmo health check

## Nota de chaves locais para E2E externo

Causa raiz observada em execucao contra chain externa:

- erro: `key 'merchant' not found`
- o E2E assina transacoes locais com chaves do keyring (`merchant` e `payer` por padrao);
- sem chave local (ou sem saldo em `ubyx`), o smoke nao chega ao `E2E_UBYX_OK`.

Correcao operacional:

- `make e2e-webhook-ubyx-keys` cria/verifica a chave `merchant` (sem imprimir seed);
- `make preflight-webhook-ubyx` agora valida:
  - backend de keyring;
  - existencia das chaves exigidas;
  - endereco publico;
  - saldo em `ubyx` para `merchant` e `payer`.

## Ordem de subida

1. Chain
2. Mock merchant
3. Relay
4. Preflight
5. Doctor (opcional)
6. E2E

## Comandos

### 1) Subir stack automatica

Modo automatico (default seguro):

```bash
make stack-webhook-ubyx-up
```

Forcando modo `ignite`:

```bash
BYX_CHAIN_MODE=ignite make stack-webhook-ubyx-up
```

Forcando modo `byxd`:

```bash
BYX_CHAIN_MODE=byxd \
BYXD_BIN=byxd \
BYX_HOME="$HOME/.byx" \
make stack-webhook-ubyx-up
```

Forcando modo `custom`:

```bash
BYX_CHAIN_MODE=custom \
BYX_CHAIN_START_CMD="byxd start --home $HOME/.byx" \
make stack-webhook-ubyx-up
```

### 2) Preflight (obrigatorio)

```bash
make preflight-webhook-ubyx
```

### 2.1) Setup rapido da chave merchant (opcional)

```bash
make e2e-webhook-ubyx-keys
```

### 3) Doctor

```bash
make doctor-webhook-ubyx
```

### 4) Rodar E2E completo

```bash
STRICT_WEBHOOK=1 make e2e-webhook-ubyx-full
```

Atalhos por modo:

```bash
make e2e-webhook-ubyx-external
make e2e-webhook-ubyx-byxd
make e2e-webhook-ubyx-custom
```

Execucao recomendada contra chain local ja ativa:

```bash
BYX_CHAIN_MODE=external \
BYX_REST=http://127.0.0.1:1317 \
BYX_RPC=http://127.0.0.1:26657 \
STRICT_WEBHOOK=1 \
make e2e-webhook-ubyx-full
```

Execucao contra endpoint remoto seguro (sem dados reais):

```bash
BYX_CHAIN_MODE=external \
BYX_REST=https://SEU_REST_SEGURO \
BYX_RPC=https://SEU_RPC_SEGURO \
STRICT_WEBHOOK=1 \
make e2e-webhook-ubyx-full
```

Acesso remoto de REST/RPC deve ficar protegido por proxy, firewall, VPN, whitelist ou outro controle equivalente.

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
- `BYX_CHAIN_MODE` (`external|byxd|custom|ignite`)
- `BYX_CHAIN_ID` (opcional)
- `BYXD_BIN` (default `byxd`, usado no modo `byxd`)
- `BYX_HOME` (opcional, usado no modo `byxd`)
- `BYX_CHAIN_START_CMD` (obrigatorio no modo `custom`)
- `LOJA_ID` (default `1`)
- `AMOUNT_UBYX` (default `500000`)
- `MERCHANT_KEY` (default `merchant`)
- `PAYER_KEY` (default `payer`)
- `KEYRING_BACKEND` (default `test`)
- `STATE_PATH` (default `.e2e/webhook-ubyx/state.json`)
- `MOCK_MERCHANT_URL` (default `http://127.0.0.1:4000/webhook`)
- `MOCK_MERCHANT_PORT` (opcional; sobrescreve porta do mock no stack-up)
- `MOCK_MERCHANT_HEALTH_URL` (opcional; default derivado para `.../health`)
- `WEBHOOK_RELAY_URL` (opcional, apenas para log)
- `MOCK_EVENTS_LOG_PATH` (default `/tmp/byx_mock_events.jsonl`)
- `MERCHANT_WEBHOOK_SECRET`
- `STRICT_WEBHOOK` (default `1`)
- `CHAIN_BOOT_TIMEOUT_S` (default `300`)
- `MIN_MERCHANT_BALANCE_UBYX` (default `1`)
- `MIN_PAYER_BALANCE_UBYX` (default `AMOUNT_UBYX`)

## Diagnostico rapido

### REST indisponivel

- erro tipico: `REST unavailable: http://127.0.0.1:1317`
- acao:

```bash
curl -sf "$BYX_REST/cosmos/base/tendermint/v1beta1/syncing"
make doctor-webhook-ubyx
```

Se o doctor/stack log mostrar:

```text
Ignite mode failed while trying to access buf.build.
This is an environment/network/proto-cache issue.
Use BYX_CHAIN_MODE=external for an already running chain,
BYX_CHAIN_MODE=byxd for a built binary,
or BYX_CHAIN_MODE=custom with BYX_CHAIN_START_CMD.
```

Entao o problema e de ambiente/startup, nao de regressao funcional do fluxo `ubyx`.

### RPC indisponivel

- acao:

```bash
curl -sf "$BYX_RPC/status"
make doctor-webhook-ubyx
```

### Relay nao envia webhook

- confirmar relay rodando com `REST_ENDPOINT`/`MERCHANT_WEBHOOK_URL` corretos
- confirmar `STATE_PATH` acessivel
- checar `.e2e/webhook-ubyx/state.json`

Validacao manual recomendada:

```bash
cd webhook-relay
npm install
npm run typecheck --if-present
npm run build --if-present
npm test --if-present
STATE_PATH="../.e2e/webhook-ubyx/state.json" npm start
ls -la ../.e2e/webhook-ubyx/state.json
```

### Chave merchant/payer ausente ou sem saldo

- erro tipico: `ERROR: key 'merchant' not found`
- acao recomendada:

```bash
make e2e-webhook-ubyx-keys
make preflight-webhook-ubyx
```

- verificar endereco publico de chave:

```bash
byxd keys show merchant -a --keyring-backend test
byxd keys show payer -a --keyring-backend test
```

- verificar saldo em `ubyx` via REST:

```bash
curl -s "$BYX_REST/cosmos/bank/v1beta1/balances/<ENDERECO>" | jq .
```

- em chain devnet externa, financiar somente com fundos de teste (sem seed real, sem dinheiro real).

### Mock merchant nao fica saudavel

- erro tipico: `mock merchant did not become healthy`
- validar manualmente:

```bash
curl -i http://127.0.0.1:4000/health
```

- esperado: HTTP `200` com JSON `{"ok":true,"service":"mock-merchant"}`
- o stack-up agora mostra:
  - URL de health testada
  - caminho de `mock-merchant.log`
  - ultimas linhas do log

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

## Observacao sobre proto

O caminho de E2E webhook/ubyx nao executa `make proto-gen`, `buf generate` ou geracao de proto.
O smoke depende apenas de chain ativa e app/binario ja compilado.

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
- `.e2e/webhook-ubyx/chain_mode.txt`
- `.e2e/webhook-ubyx/env_summary.txt` (sem segredos)
- `.e2e/webhook-ubyx/startup_command.txt` (mascarado)
- `.e2e/webhook-ubyx/failure_reason.txt` (quando houver falha)
