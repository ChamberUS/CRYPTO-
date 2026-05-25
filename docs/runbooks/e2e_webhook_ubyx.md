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

## Nota de compatibilidade Bash macOS

Causa raiz observada em ambiente macOS:

- o script usava `mapfile` para montar `tx_args`;
- o Bash padrao do macOS (3.2) nao possui `mapfile`;
- erro observado: `mapfile: command not found` seguido de `tx_args[0]: unbound variable`.

Correcao aplicada:

- `mapfile` foi substituido por loop compatível:
  - `while IFS= read -r arg; do ... done`
- quando nao for possivel montar argumentos de tx, o script falha com:
  - `failed to build tx args`

## Nota de compatibilidade da CLI `byxd` (`--amount-ubyx`)

Causa raiz observada em execucao recente:

- erro: `unknown flag: --amount-ubyx`
- o host estava usando binario `byxd` antigo, ainda com interface legada (`amount_microbyx`).

Correcao aplicada:

- `x/payments/module/autocli.go` foi ajustado para alinhar parse real da CLI:
  - `create-payment-request` agora declara `PositionalArgs` para `loja_id` e `amount` no runtime AutoCLI;
  - para compatibilidade com o descriptor carregado no runtime atual, o segundo positional usa alias interno legado (`amount_microbyx`) apenas como mapeamento de parser;
  - o comando principal continua exposto e usado como `create-payment-request [loja-id] [amount-ubyx]`.
- `scripts/preflight_webhook_ubyx.sh` agora valida automaticamente:
  - qual binario foi resolvido em `BYXD_BIN`;
  - path efetivo do binario;
  - versao/build info (quando disponivel);
  - suporte da CLI a assinatura posicional `[loja-id] [amount-ubyx]`.
  - parse runtime de positional args com probe `--generate-only` (sem `--offline`) para detectar o caso:
    - help mostra `[amount-ubyx]`, mas runtime rejeita com `accepts 0 arg(s), received 2`.
- `scripts/preflight_webhook_ubyx.sh` valida tambem que o script E2E nao usa:
  - `--amount-ubyx`
  - `--amount-microbyx`
  - validacao feita no bloco executavel do comando `create-payment-request` (nao em mensagens/strings de diagnostico).
- `scripts/e2e_payments_webhook_ubyx.sh` usa somente chamada posicional:
  - `create-payment-request "$LOJA_ID" "$AMOUNT_UBYX"`.
  - mensagem de erro alinhada para: `byxd CLI does not expose positional [amount-ubyx]`.
- o E2E agora usa `BYXD_BIN` (default `byxd`) em todas as chamadas de CLI.

Correcao adicional de preflight:

- causa raiz de falha recente: probe usava combinacao instavel/invalida com `--offline + --generate-only` (e podia conflitar com chain-id de config local);
- ajuste: remover `--offline` do probe e manter somente `--generate-only` com contexto minimo (`--fees`, `--gas`, `--account-number`, `--sequence`);
- fallback: se o probe ainda falhar por variação de SDK/config local, ele vira `WARN` (não bloqueante);
- em falha do probe, o preflight agora imprime:
  - comando usado;
  - stderr capturado;
  - hint explicito sobre combinacoes instaveis.

Validacao manual:

```bash
./bin/byxd tx payments create-payment-request --help | grep -i amount
```

Se a usage nao mostrar `[amount-ubyx]`, rebuild local:

```bash
mkdir -p bin
go build -o ./bin/byxd ./cmd/byxd
BYXD_BIN=./bin/byxd BYX_CHAIN_MODE=external make e2e-webhook-ubyx-full
```

Diagnostico adicional (parse real):

```bash
./bin/byxd tx payments create-payment-request 1 500000 --help
```

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
BYXD_BIN=./bin/byxd \
BYX_CHAIN_MODE=external \
BYX_REST=http://127.0.0.1:1317 \
BYX_RPC=http://127.0.0.1:26657 \
STRICT_WEBHOOK=1 \
make e2e-webhook-ubyx-full
```

Execucao contra endpoint remoto seguro (sem dados reais):

```bash
BYXD_BIN=./bin/byxd \
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

Artefatos relevantes da criacao da payment request:

- `create_payment_request_broadcast.json`
- `create_payment_request_broadcast.stderr.txt`
- `create_payment_request_broadcast.raw.txt`
- `create_payment_request_tx.json`
- `txhash_create_payment_request.txt`
- `create_payment_request_command.txt`
- `e2e_memo.txt`
- `payment_request_query.json`
- `payment_request_query.http.txt`
- `payment_request_qr.json`
- `payment_request_qr.http.txt`
- `merchant_query_by_id.json`
- `merchant_query_all.json`
- `create_merchant_broadcast.json`
- `create_merchant_broadcast.stderr.txt`
- `create_merchant_broadcast.raw.txt`
- `create_merchant_tx.json`
- `pay_request_broadcast.json`
- `pay_request_broadcast.stderr.txt`
- `pay_request_broadcast.raw.txt`
- `pay_request_tx.json`
- `pay_request_command.txt`
- `merchant_id.txt`
- `merchant_signer_info.txt`
- `merchant_account_onchain.json`
- `chain_status.json`
- `request_id.txt`
- `wait_tx_last_response.txt`

### 6) Derrubar stack (manual)

```bash
make stack-webhook-ubyx-down
```

## Variaveis de ambiente

- `BYX_REST` (default `http://127.0.0.1:1317`)
- `BYX_RPC` (default `http://127.0.0.1:26657`)
- `BYX_CHAIN_MODE` (`external|byxd|custom|ignite`)
- `BYX_CHAIN_ID` (opcional)
- `BYXD_BIN` (default `byxd`, usado no `preflight`, no `e2e` e no modo `byxd`)
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

### Merchant no keyring vs merchant no modulo `lojas`

- a chave `merchant` no keyring serve apenas para assinar a tx;
- `payments` nao usa a chave diretamente para resolver `loja_id`;
- `payments` chama `lojasKeeper.GetMerchant(ctx, loja_id)`, entao o `loja_id` precisa existir como `Merchant` on-chain no modulo `lojas`;
- por isso e possivel ter:
  - chave local existente;
  - saldo on-chain existente;
  - e ainda assim falhar com `merchant not found`.

Consultas uteis:

```bash
./bin/byxd query lojas merchant 1 --node http://127.0.0.1:26657 -o json | jq
./bin/byxd query lojas merchant-all --node http://127.0.0.1:26657 -o json | jq
```

Criacao manual de merchant/loja:

```bash
MERCHANT_ADDR="$(./bin/byxd keys show merchant -a --keyring-backend test)"
./bin/byxd tx lojas create-merchant \
  "BYX E2E Merchant" \
  "BYX E2E Address" \
  "$MERCHANT_ADDR" \
  "e2e" \
  "0000000000000000000000000000000000000000000000000000000000000000" \
  "pending" \
  --from merchant \
  --keyring-backend test \
  --chain-id byx-devnet-private-1 \
  --node http://127.0.0.1:26657 \
  --fees 0ubyx \
  --gas auto \
  --gas-adjustment 1.3 \
  --broadcast-mode sync \
  --yes \
  --output json
```

Observacao de robustez do E2E:

- o broadcast de `create-merchant` agora salva:
  - stdout JSON em `create_merchant_broadcast.json`
  - stderr puro em `create_merchant_broadcast.stderr.txt`
- stdout bruto em `create_merchant_broadcast.raw.txt`
- stderr nunca mais e gravado dentro de `.json`;
- se o stdout nao for JSON valido, o script imprime stdout + stderr e falha com erro claro.
- o mesmo padrao agora vale para:
  - `create-payment-request`
  - `pay-payment-request`

### `invalid json` em broadcasts

- causa observada: o Cosmos SDK pode imprimir linhas como `gas estimate: ...` no stdout antes do JSON final do broadcast;
- isso quebra `jq` quando o arquivo `.json` recebe a saida bruta inteira;
- o E2E agora usa um helper unico para txs:
  - salva stdout bruto em `.raw.txt`;
  - salva stderr em `.stderr.txt`;
  - extrai o ultimo objeto JSON valido para o `.json`;
  - falha com `returned invalid json` se nenhum JSON valido for encontrado;
  - so chama `wait_tx` quando `code == 0`.

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

### `tx not indexed` na criacao da request

- causa raiz observada: o broadcast `sync` retornava `txhash`, mas o RPC ainda nao tinha indexado a tx;
- o fluxo antigo podia falhar cedo e, em seguida, chamar `jq` em arquivo vazio/inexistente;
- o fluxo atual espera a indexacao com retry por ate `120s` (`WAIT_TX_ATTEMPTS=60`, `WAIT_TX_SLEEP_S=2`);
- em timeout, inspecionar `.e2e/webhook-ubyx/wait_tx_last_response.txt` e confirmar `tx_index`.

Diagnostico manual:

```bash
./bin/byxd q tx <TXHASH> --node http://127.0.0.1:26657 -o json | jq
curl -s http://127.0.0.1:26657/status | jq -r '.result.node_info.other.tx_index'
```

- se a tx indexar e ainda assim o `request_id` nao for resolvido, o smoke imprime os eventos disponiveis e falha com `could not resolve request_id from indexed tx events`.

### `code=4` na criacao da request

- `txhash` no retorno de broadcast nao significa que a tx foi aceita em bloco;
- se `create_payment_request_broadcast.json` vier com `code != 0`, a tx foi rejeitada no `CheckTx`/antehandler;
- neste caso o smoke falha imediatamente com `create-payment-request broadcast rejected`;
- o fluxo nao chama `wait_tx` nem `q tx`, porque `tx not found` passa a ser consequencia da rejeicao, nao de indexacao.

Diagnostico manual:

```bash
./bin/byxd keys show merchant -a --keyring-backend test
./bin/byxd q auth account <MERCHANT_ADDR> --node http://127.0.0.1:26657 -o json | jq
curl -s http://127.0.0.1:1317/cosmos/auth/v1beta1/accounts/<MERCHANT_ADDR> | jq
curl -s http://127.0.0.1:26657/status | jq -r '.result.node_info.network'
```

Comando manual equivalente:

```bash
./bin/byxd tx payments create-payment-request 1 500000 \
  --from merchant \
  --keyring-backend test \
  --chain-id byx-devnet-private-1 \
  --node http://127.0.0.1:26657 \
  --fees 0ubyx \
  --gas auto \
  --gas-adjustment 1.3 \
  --broadcast-mode sync \
  --yes \
  --output json
```

- se o retorno trouxer `code: 4` e `signature verification failed`, comparar `account_number`, `sequence`, `chain-id` e o endereco atual da chave `merchant` com a conta financiada.

### `merchant not found` na criacao da request

- se a tx falhar com `merchant not found: key not found`, o problema e de pre-condicao funcional do modulo `lojas`;
- o `LOJA_ID` informado nao existe como `Merchant` on-chain, ou existe mas pertence a outro `creator`;
- o E2E agora executa `ensure_merchant_or_loja()` antes do `create-payment-request`:
  - consulta `merchant` por `LOJA_ID`;
  - valida se `merchant.creator` bate com o endereco da chave `merchant`;
  - se nao bater, procura outro `merchant` do mesmo `creator`;
  - se ainda nao existir, faz `tx lojas create-merchant`, aguarda a tx e resolve o `merchant_id` real;
  - atualiza o `LOJA_ID` usado no `create-payment-request`.

- se a pre-condicao nao puder ser garantida, o erro agora e explicito:
  - `merchant/loja prerequisite missing`

### `created` vs `reused` em payment request

- `CreatePaymentRequest` tem dois caminhos validos no modulo `payments`:
  - cria nova request e emite `byx_payment_request_created`
  - reutiliza request pendente equivalente e emite `byx_payment_request_reused`
- a regra de dedupe usa:
  - `loja_id`
  - `amount_ubyx`
  - `memo`
- por isso o E2E agora define um `E2E_MEMO` unico por execucao e salva em `e2e_memo.txt`;
- a validacao do fluxo aceita `amount_ubyx` e `request_id` tanto em `created` quanto em `reused`.

Observacao operacional:

- `pay-payment-request` e posicional no AutoCLI:
  - `pay-payment-request [request-id]`
- o E2E nao usa mais `--request-id`.

### Falha silenciosa apos `REQUEST_ID`

- antes, `assert_query_fields` usava `curl -sf` diretamente em command substitution;
- com `set -e`, qualquer erro HTTP podia encerrar o script sem contexto claro logo apos `REQUEST_ID=<id>`;
- o E2E agora usa `curl_json_checked(label, url, out_file, http_file)` para:
  - salvar body em arquivo;
  - salvar HTTP code em arquivo;
  - imprimir `label`, `url`, `http_code` e `body` quando o endpoint falha;
  - validar JSON com `jq` antes de seguir.
- `assert_query_fields` agora emite marcadores explicitos:
  - `ASSERT_QUERY_FIELDS_START request_id=...`
  - `ASSERT_QUERY_FIELDS_OK request_id=...`
- `pay_request` agora emite:
  - `PAY_REQUEST_START request_id=...`
  - `PAY_REQUEST_DONE request_id=...`
- antes do broadcast do pagamento, o script grava `pay_request_command.txt` para confirmar entrada nessa etapa.

Artefatos dessa etapa:

- `payment_request_query.json`
- `payment_request_query.http.txt`
- `payment_request_qr.json`
- `payment_request_qr.http.txt`

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

Comando recomendado para o proximo ciclo E2E:

```bash
BYXD_BIN=./bin/byxd \
BYX_CHAIN_MODE=external \
BYX_REST=http://127.0.0.1:1317 \
BYX_RPC=http://127.0.0.1:26657 \
STRICT_WEBHOOK=1 \
make e2e-webhook-ubyx-full
```

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
- `.e2e/webhook-ubyx/create_payment_request_broadcast.json`
- `.e2e/webhook-ubyx/create_payment_request_broadcast.stderr.txt`
- `.e2e/webhook-ubyx/create_payment_request_broadcast.raw.txt`
- `.e2e/webhook-ubyx/create_payment_request_tx.json`
- `.e2e/webhook-ubyx/txhash_create_payment_request.txt`
- `.e2e/webhook-ubyx/create_payment_request_command.txt`
- `.e2e/webhook-ubyx/e2e_memo.txt`
- `.e2e/webhook-ubyx/payment_request_query.json`
- `.e2e/webhook-ubyx/payment_request_query.http.txt`
- `.e2e/webhook-ubyx/payment_request_qr.json`
- `.e2e/webhook-ubyx/payment_request_qr.http.txt`
- `.e2e/webhook-ubyx/merchant_query_by_id.json`
- `.e2e/webhook-ubyx/merchant_query_all.json`
- `.e2e/webhook-ubyx/create_merchant_broadcast.json`
- `.e2e/webhook-ubyx/create_merchant_broadcast.stderr.txt`
- `.e2e/webhook-ubyx/create_merchant_broadcast.raw.txt`
- `.e2e/webhook-ubyx/create_merchant_tx.json`
- `.e2e/webhook-ubyx/pay_request_broadcast.json`
- `.e2e/webhook-ubyx/pay_request_broadcast.stderr.txt`
- `.e2e/webhook-ubyx/pay_request_broadcast.raw.txt`
- `.e2e/webhook-ubyx/pay_request_tx.json`
- `.e2e/webhook-ubyx/pay_request_command.txt`
- `.e2e/webhook-ubyx/merchant_id.txt`
- `.e2e/webhook-ubyx/merchant_signer_info.txt`
- `.e2e/webhook-ubyx/merchant_account_onchain.json`
- `.e2e/webhook-ubyx/chain_status.json`
- `.e2e/webhook-ubyx/request_id.txt`
- `.e2e/webhook-ubyx/wait_tx_last_response.txt`
- `.e2e/webhook-ubyx/chain_mode.txt`
- `.e2e/webhook-ubyx/env_summary.txt` (sem segredos)
- `.e2e/webhook-ubyx/startup_command.txt` (mascarado)
- `.e2e/webhook-ubyx/failure_reason.txt` (quando houver falha)

## Confirmacao real do E2E

Execucao validada em ambiente real:

- marcador final observado: `E2E_UBYX_OK request_id=6`
- payload de pagamento observado:
  - `loja_id=1`
  - `amount_ubyx=500000`
- payload final de webhook observado:
  - `request_id=6`
  - `loja_id=1`
  - `amount_ubyx=500000`
  - `paid_at_unix=<unix>`
- marcadores de progresso observados:
  - `PAY_REQUEST_START`
  - `PAY_REQUEST_DONE`
- estado do webhook confirmado em `state.json`

Observacao:

- os artefatos em `.e2e/webhook-ubyx/` sao apenas locais de diagnostico;
- `.e2e/` deve permanecer fora do commit.
