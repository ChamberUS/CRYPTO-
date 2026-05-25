# Sprint 8 - Status atual

Data: 2026-05-22

## Resumo simples

A BYX continua em fase de endurecimento técnico antes de qualquer uso com dinheiro real.

Nesta etapa:

- a unidade monetária da blockchain segue consolidada em `ubyx`;
- o sistema continua com supply fixo;
- cashback, faucet e payments continuam usando reserva existente, sem criação dinâmica de moedas;
- os campos públicos de protobuf e API foram migrados de nomes legados com `microBYX` para nomes com `ubyx`.

## Decisão monetária

A regra oficial é:

- unidade base on-chain: `ubyx`
- unidade exibida para pessoas: `BYX`
- conversão: `1 BYX = 1_000_000 ubyx`

Isso evita ambiguidade e permite cashback, taxas e microtransações com precisão.

## Supply cap atual

O teto de supply continua fixo em:

- `1_000_000_000 BYX`
- equivalente a `1_000_000_000_000_000 ubyx`

## Objetivo desta migração

O objetivo desta tarefa foi remover a ambiguidade pública dos payloads econômicos da devnet.

Antes desta mudança, a API pública ainda expunha nomes como `amount_microbyx` e `cashback_micro_byx`, mesmo após a adoção oficial de `ubyx` como unidade mínima on-chain.

Agora, os campos públicos foram renomeados para refletir a unidade correta.

## Campos renomeados

No módulo `lojas`:

- `cashback_rate_micro_byx_per_real` -> `cashback_rate_ubyx_per_real`
- `max_cashback_micro_byx_por_venda` -> `max_cashback_ubyx_por_venda`
- `max_cashback_daily_per_loja_microbyx` -> `max_cashback_daily_per_loja_ubyx`
- `cashback_micro_byx` -> `cashback_ubyx`
- `total_cashback_micro_byx` -> `total_cashback_ubyx`

No módulo `payments`:

- `amount_microbyx` -> `amount_ubyx`

## O que foi feito nesta tarefa

- os arquivos `.proto` foram atualizados mantendo os mesmos field numbers;
- os tipos Go gerados, artefatos TypeScript e OpenAPI/Swagger foram alinhados com os novos nomes;
- os eventos e testes passaram a usar os nomes públicos com `ubyx`;
- backend devnet, webhook relay e mock merchant foram atualizados para o novo payload;
- `TransferBYXFromReserve` continua sendo o caminho principal para distribuição econômica;
- `MintBYXTo` continua apenas como compatibilidade legada/deprecated, sem voltar a ser o fluxo principal.

## Breaking changes

Esta mudança é breaking para consumidores da devnet que ainda esperavam os nomes antigos de payload:

- `amount_microbyx`
- `cashback_micro_byx`
- `total_cashback_micro_byx`
- `cashback_rate_micro_byx_per_real`
- `max_cashback_micro_byx_por_venda`
- `max_cashback_daily_per_loja_microbyx`

Quem consome a API da devnet deve passar a usar os nomes com `ubyx`.

Não houve mudança em:

- namespace protobuf `/byx/...`
- prefixo bech32 `byx`
- nome da chain/app `byx`
- display humano `BYX`

## Arquivos alterados

Grupos principais alterados nesta tarefa:

- protos em `proto/byx/lojas/v1/` e `proto/byx/payments/v1/`
- tipos e keepers em `x/lojas/` e `x/payments/`
- testes de app e módulos
- backend fechado em `integrations/byx-backend/`
- relay e mock merchant em `webhook-relay/`
- artefatos gerados em `x/.../*.pb.go`, `byx/...`, `docs/static/openapi.json`

## Geração de protobuf e artefatos

Foi tentado primeiro:

- `make proto-gen`

O comando ficou travado no `ignite generate proto-go --yes` no ambiente atual e não concluiu de forma útil.

Depois foi usado o caminho equivalente com `buf`:

- `buf generate --template proto/buf.gen.gogo.yaml`
- `buf generate --template proto/buf.gen.sta.yaml`
- `buf generate --template proto/buf.gen.swagger.yaml`
- `buf generate --template proto/buf.gen.ts.yaml`

Durante isso foi identificado um problema no template `proto/buf.gen.sta.yaml`:

- a flag `ignore_comments=true` não era aceita pela versão local de `protoc-gen-openapiv2`

Solução aplicada:

- remover a flag inválida do template e rerodar o `buf generate`

## Testes executados

Foram executados com sucesso:

- `GOCACHE=$(pwd)/.gocache go test ./x/lojas/... -count=1`
- `GOCACHE=$(pwd)/.gocache go test ./x/payments/... -count=1`
- `GOCACHE=$(pwd)/.gocache go test ./app -count=1`
- `node --check integrations/byx-backend/src/server.js`

## Pendências restantes

Ainda existem ocorrências de `byx` que estão corretas e devem permanecer quando significam:

- nome da chain/app
- prefixo bech32
- namespace protobuf
- rotas `/byx/...`

Também ainda existem usos residuais de `MintCoins` em código de bootstrap ou testes específicos, fora do runtime econômico normal.

## Próximo passo recomendado

O próximo passo técnico recomendado é endurecer os consumidores externos da devnet:

1. revisar contratos JSON dos integradores fechados;
2. validar novamente smoke tests de webhook e payment request com payload `amount_ubyx`;
3. depois seguir para o ramp sandbox, sem gateway real.

## Validação E2E fechada (2026-05-23)

### Objetivo

Validar ponta a ponta o fluxo Payment Request -> Pay -> Evento -> Webhook Relay -> Mock Merchant com os payloads novos em `ubyx`.

### Scripts usados/criados

- script novo: `scripts/e2e_payments_webhook_ubyx.sh`
- script legado mantido: `scripts/e2e_payments_webhook.sh`

### Payloads usados

- criação de pedido: `{"loja_id": <id>, "amount_ubyx": <valor>}`
- webhook para merchant: `{"request_id":..., "loja_id":..., "amount_ubyx":..., "paid_at_unix":..., "event_id":..., "trace_id":...}`

### Validações implementadas no smoke ubyx

- chain/REST acessíveis
- create request com `--amount-ubyx`
- query e QR com `amount_ubyx`
- ausência de `amount_microbyx` no fluxo ativo
- pagamento e status `PAYMENT_STATUS_PAID`
- evento de criação contendo `amount_ubyx`
- estado de entrega do relay em `state.json`
- replay com mesmo `X-BYX-Idempotency-Key` e `X-BYX-Event-Id` (esperado: dedupe)
- assinatura inválida (esperado: HTTP 401)

### Correções feitas

- mock merchant agora registra eventos de teste em JSONL opcional (`MOCK_EVENTS_LOG_PATH`) para assert de:
  - entrega válida (`status=ok`)
  - duplicata (`status=duplicate`)
  - assinatura inválida (`status=invalid_signature`)
- docs E2E atualizadas para apontar o smoke `..._ubyx.sh`

### Resultado

- checks locais de compilação/testes passaram
- execução do smoke depende de stack local ativa (chain + relay + mock), conforme pré-requisitos abaixo

### Pré-requisitos para execução completa do smoke

1. subir chain: `ignite chain serve --reset-once`
2. subir mock merchant:
   - `cd webhook-relay/mock-merchant`
   - `MOCK_EVENTS_LOG_PATH=/tmp/byx_mock_events.jsonl MERCHANT_WEBHOOK_SECRET=devsecret PORT=4000 npm start`
3. subir relay:
   - `cd webhook-relay`
   - `REST_ENDPOINT=http://127.0.0.1:1317 LOJA_ID=1 MERCHANT_WEBHOOK_URL=http://127.0.0.1:4000/webhook MERCHANT_WEBHOOK_SECRET=devsecret STATE_PATH=./state.json npm start`
4. rodar smoke:
   - `STRICT_WEBHOOK=1 STATE_PATH=./webhook-relay/state.json MOCK_EVENTS_LOG_PATH=/tmp/byx_mock_events.jsonl MERCHANT_WEBHOOK_SECRET=devsecret ./scripts/e2e_payments_webhook_ubyx.sh`

### Pendências

- executar o smoke completo em ambiente com stack local efetivamente ativa e anexar logs de execução no próximo ciclo.

### Próximo passo recomendado

- rodar o smoke `e2e_payments_webhook_ubyx.sh` em CI/runner com serviços auxiliares (chain, relay, mock) e registrar artefatos (`state.json`, log JSONL do mock).

### Execução local deste ciclo (2026-05-23)

Comando executado:

- `bash scripts/e2e_payments_webhook_ubyx.sh`

Resultado:

- falhou no pré-check de infraestrutura com: `REST unavailable: http://127.0.0.1:1317`

Interpretação:

- o script está validando corretamente os pré-requisitos;
- para validação E2E completa é necessário subir a stack local (chain + relay + mock) antes do smoke.

## Padronizacao do E2E webhook/ubyx (2026-05-23)

### Objetivo

Reduzir dependencia de contexto manual para qualquer pessoa/agent rodar a stack local e o smoke E2E webhook com payload `ubyx`.

### Alvos Makefile criados

- `make preflight-webhook-ubyx`
- `make doctor-webhook-ubyx`
- `make e2e-webhook-ubyx`

### Pre-requisitos validados

- `byxd`, `curl`, `jq`, `openssl`, `node`, `npm`
- REST em `BYX_REST` (default `http://127.0.0.1:1317`)
- RPC em `BYX_RPC` (default `http://127.0.0.1:26657`)

### Motivo da falha anterior

A falha anterior do smoke foi de ambiente/infra local nao iniciada:

- `REST unavailable: http://127.0.0.1:1317`

Nao indica regressao obrigatoria na logica de `payments`/`webhook`.

### Arquivos de operacionalizacao

- `scripts/preflight_webhook_ubyx.sh`
- `scripts/e2e_payments_webhook_ubyx.sh`
- `docs/runbooks/e2e_webhook_ubyx.md`

### Proximo passo

1. executar o smoke completo com chain + relay + mock ativos
2. anexar logs/artefatos (`state.json`, JSONL do mock)
3. com smoke verde, avancar para ramp sandbox Pix/BYX (sem gateway real)

## Padronizacao da subida de stack local (2026-05-23)

### Objetivo

Padronizar uma forma reproduzivel para qualquer pessoa subir stack local (chain + relay + mock), rodar preflight/doctor/e2e e coletar artefatos.

### Scripts criados/atualizados

- `scripts/e2e_webhook_ubyx_stack_up.sh`
- `scripts/e2e_webhook_ubyx_stack_down.sh`
- `scripts/e2e_webhook_ubyx_collect_artifacts.sh`
- `scripts/e2e_payments_webhook_ubyx.sh` (mensagens de erro e variaveis de ambiente)
- `scripts/preflight_webhook_ubyx.sh`

### Alvos Makefile criados/atualizados

- `make stack-webhook-ubyx-up`
- `make stack-webhook-ubyx-down`
- `make stack-webhook-ubyx-logs`
- `make e2e-webhook-ubyx-full`
- `make preflight-webhook-ubyx` (agora grava log e falha corretamente)
- `make doctor-webhook-ubyx` (diagnostico mais explicito)
- `make e2e-webhook-ubyx` (agora grava log e falha corretamente)

### Resultado da execucao neste ciclo

Comando executado:

- `make e2e-webhook-ubyx-full`

Resultado:

- **falhou na subida da chain**, antes de chegar no pagamento/webhook.

Causa exata (em `.e2e/webhook-ubyx/chain.log`):

- `ignite chain serve --reset-once` nao conseguiu finalizar build por indisponibilidade de acesso remoto ao `buf.build` durante etapa de proto.

Impacto:

- REST `1317` e RPC `26657` nao ficaram disponiveis;
- preflight/doctor/e2e confirmam falha de ambiente;
- smoke funcional completo permanece pendente de ambiente com chain ativa.

### Evidencias e artefatos

Diretorio:

- `.e2e/webhook-ubyx/`

Arquivos principais:

- `chain.log`
- `preflight.log`
- `doctor.log`
- `artifacts.txt`

### Ajuste operacional aplicado

- `stack_up` agora usa `HOME` local em `.e2e/webhook-ubyx/home` para evitar erro de permissao em `~/.ignite`.
- timeout de bootstrap da chain foi aumentado para `300s` (`CHAIN_BOOT_TIMEOUT_S`).

### Pendencias

- executar novamente o smoke em ambiente com acesso necessario para startup da chain (ou com comando alternativo de chain preprovisionada via `BYX_CHAIN_START_CMD`).
- obter saida com `E2E_UBYX_OK`.

### Proximo passo recomendado

1. subir chain por comando funcional no ambiente alvo (com rede para `buf.build` ou alternativa local preprovisionada);
2. executar `STRICT_WEBHOOK=1 make e2e-webhook-ubyx-full`;
3. confirmar `E2E_UBYX_OK` e anexar `e2e.log`, `state.json` e `mock-events.jsonl`.

## Desacoplamento do E2E de `ignite` (2026-05-24)

### Objetivo

Remover dependencia obrigatoria de `ignite chain serve --reset-once` no caminho do smoke E2E webhook/ubyx.

### Modos implementados (`BYX_CHAIN_MODE`)

- `external`: usa chain ja ativa (nao tenta subir chain).
- `byxd`: sobe com binario local (`BYXD_BIN`, default `byxd`).
- `custom`: sobe com comando definido em `BYX_CHAIN_START_CMD`.
- `ignite`: mantem fluxo de `ignite chain serve --reset-once` de forma opcional.

Default seguro aplicado:

- se `BYX_CHAIN_MODE` nao estiver setado e REST/RPC estiverem ativos, usa `external`;
- se REST/RPC nao estiverem ativos, falha com orientacao explicita para escolher `external|byxd|custom|ignite`.

### Scripts/targets atualizados

- `scripts/e2e_webhook_ubyx_stack_up.sh`
- `scripts/preflight_webhook_ubyx.sh`
- `scripts/doctor_webhook_ubyx.sh` (novo)
- `scripts/e2e_webhook_ubyx_collect_artifacts.sh`
- `scripts/e2e_payments_webhook_ubyx.sh`
- `Makefile`
- `docs/runbooks/e2e_webhook_ubyx.md`

Novos alvos:

- `make e2e-webhook-ubyx-external`
- `make e2e-webhook-ubyx-byxd`
- `make e2e-webhook-ubyx-custom`

### Diagnostico de falha `buf.build`

Quando `BYX_CHAIN_MODE=ignite` falha por acesso/proto-cache, os logs agora exibem:

- `Ignite mode failed while trying to access buf.build.`
- `This is an environment/network/proto-cache issue.`
- `Use BYX_CHAIN_MODE=external for an already running chain,`
- `BYX_CHAIN_MODE=byxd for a built binary,`
- `or BYX_CHAIN_MODE=custom with BYX_CHAIN_START_CMD.`

### Artefatos adicionais

Diretorio `.e2e/webhook-ubyx/` agora inclui:

- `chain_mode.txt`
- `env_summary.txt` (sem segredos)
- `startup_command.txt` (mascarado)
- `failure_reason.txt` (quando houver falha)

### Estado funcional/economico

- `ubyx` permanece como unidade base on-chain.
- payload publico continua em `amount_ubyx`/`cashback_ubyx`.
- nao houve mudanca de namespace `/byx/...`, bech32 `byx`, app/chain `byx`, display `BYX`.
- nao houve reintroducao de mint dinamico em runtime economico.

### Proximo passo recomendado

1. executar `BYX_CHAIN_MODE=external STRICT_WEBHOOK=1 make e2e-webhook-ubyx-full` em ambiente com chain ativa;
2. capturar evidencia com `E2E_UBYX_OK`;
3. avancar para ramp sandbox Pix/BYX (sem dinheiro real).

## Correcao do health check do mock merchant (2026-05-25)

### Causa raiz confirmada

- chain externa podia estar saudavel em `1317/26657`, mas o stack-up falhava em `mock merchant did not become healthy`;
- o mock merchant nao tinha endpoint de saude HTTP;
- health check antigo usava URL de webhook (`/webhook`) com `GET`, recebendo `404`.

### Correcao aplicada

- `webhook-relay/mock-merchant/server.js` agora expoe:
  - `GET /health` -> `200 {"ok":true,"service":"mock-merchant"}`
  - `GET /healthz` -> alias
- `scripts/e2e_webhook_ubyx_stack_up.sh` agora:
  - valida saude do mock em `.../health` (ou `MOCK_MERCHANT_HEALTH_URL`);
  - respeita `MOCK_MERCHANT_PORT`/porta derivada de `MOCK_MERCHANT_URL`;
  - em falha, imprime URL testada, caminho do log, tail do log e comando manual de validacao.

### Validacao manual desta tarefa

- `node --check webhook-relay/mock-merchant/server.js`: ok
- `curl -i http://127.0.0.1:4000/health`: retornou `HTTP/1.1 200 OK`
- body: `{"ok":true,"service":"mock-merchant"}`

### Resultado do E2E neste ambiente

Comando:

- `BYX_CHAIN_MODE=external BYX_REST=http://127.0.0.1:1317 BYX_RPC=http://127.0.0.1:26657 STRICT_WEBHOOK=1 make e2e-webhook-ubyx-full`

Resultado:

- falha antes do mock por indisponibilidade local de REST/RPC (`external mode requires an already running chain ...`);
- portanto `E2E_UBYX_OK` ainda nao apareceu nesta execucao local.

### Proximo passo recomendado

1. manter tunel/chain externa ativa e reexecutar o comando E2E;
2. confirmar que o stack-up passa do health check do mock;
3. capturar evidencia final com `E2E_UBYX_OK`.

## Correcao do bootstrap do webhook relay (2026-05-25)

### Causa raiz confirmada

- com chain externa e mock saudaveis, o fluxo passou a falhar na subida do relay;
- startup antigo era `node --loader ts-node/esm index.ts`;
- em Node 20 no ambiente atual, o processo abortava com warning/exception antes do loop de polling;
- `state.json` nao era criado e o stack-up encerrava com `webhook relay did not bootstrap state file`.

### Correcao aplicada

- `webhook-relay/package.json`:
  - `start` alterado para `tsx index.ts`;
  - `typecheck` padronizado para `tsc --noEmit`;
  - `tsx` adicionado em `devDependencies`.
- `webhook-relay/index.ts`:
  - cria `dirname(STATE_PATH)` com `mkdir -p` antes de ler/gravar estado.
- `scripts/e2e_webhook_ubyx_stack_up.sh`:
  - `STATE_PATH` padrao do fluxo E2E passou para `.e2e/webhook-ubyx/state.json`;
  - cria diretório pai de `STATE_PATH` antes de subir relay;
  - em falha do relay, imprime comando usado, `STATE_PATH`, tail do log e aviso de `state.json` ausente.
- `scripts/e2e_payments_webhook_ubyx.sh`:
  - default de `STATE_PATH` alinhado para `.e2e/webhook-ubyx/state.json`.
- `Makefile`:
  - `stack-webhook-ubyx-up`, `e2e-webhook-ubyx` e coleta de artefatos alinhados para o mesmo `STATE_PATH`.

### Validacao manual desta tarefa

- `STATE_PATH="../.e2e/webhook-ubyx/state.json" npm start` cria/bootstrapa o arquivo de estado no path esperado.

### Resultado do E2E neste ambiente

- comando alvo permanece:
  - `BYX_CHAIN_MODE=external BYX_REST=http://127.0.0.1:1317 BYX_RPC=http://127.0.0.1:26657 STRICT_WEBHOOK=1 make e2e-webhook-ubyx-full`
- sucesso final depende da disponibilidade efetiva do tunel/chain externa no momento da execucao local.

### Proximo passo recomendado

1. manter tunel/chain externa ativa;
2. rerodar `make e2e-webhook-ubyx-full`;
3. capturar `E2E_UBYX_OK` com artefatos `state.json`, `e2e.log` e `mock-events.jsonl`.

## Correcao de diagnostico de keyring para E2E externo (2026-05-25)

### Causa raiz confirmada

- com chain externa, mock e relay saudaveis, o fluxo passou a falhar na primeira tx com:
  - `ERROR: key 'merchant' not found`
- o smoke depende de chaves locais para assinatura (`merchant` e `payer`, por padrao);
- no ambiente local, a chave nao existia no keyring usado.

### Correcao aplicada

- `scripts/preflight_webhook_ubyx.sh` agora valida:
  - `KEYRING_BACKEND`;
  - chaves exigidas (`MERCHANT_KEY` e `PAYER_KEY`);
  - endereco publico de cada chave (sem seed);
  - saldo em `ubyx` via REST para cada endereco;
  - falha clara quando saldo for insuficiente.
- script novo: `scripts/e2e_webhook_ubyx_keys_setup.sh`
  - cria `merchant` apenas se nao existir;
  - nao imprime seed phrase;
  - nao exporta private key;
  - mostra endereco e saldo atual em `ubyx`.
- alvo novo no Makefile:
  - `make e2e-webhook-ubyx-keys`

### Resultado desta etapa

- diagnostico de chaves/saldo ficou explicito antes de tentar transacao;
- caminho de remediation local ficou reproduzivel sem dados sensiveis.

### Proximo passo recomendado

1. executar `make e2e-webhook-ubyx-keys`;
2. garantir saldo de teste em `ubyx` para `merchant` e `payer` na devnet;
3. rerodar `BYX_CHAIN_MODE=external ... make e2e-webhook-ubyx-full` ate obter `E2E_UBYX_OK`.
