# E2E webhook/ubyx stability checkpoint

## Referencia

- commit de referencia: `7d91eb8`
- branch de referencia: `main`
- marcador funcional confirmado: `E2E_UBYX_OK request_id=6`

## Comando oficial

```bash
BYXD_BIN=./bin/byxd \
BYX_CHAIN_MODE=external \
BYX_REST=http://127.0.0.1:1317 \
BYX_RPC=http://127.0.0.1:26657 \
STRICT_WEBHOOK=1 \
make e2e-webhook-ubyx-full
```

## Pre-condicoes

- chain localweb ativa
- REST `1317` acessivel
- RPC `26657` acessivel
- `tx_index` habilitado
- devnet configurada em `ubyx`
- chaves `merchant` e `payer` presentes no keyring local
- `merchant` e `payer` com saldo suficiente em `ubyx`
- mock merchant ativo
- webhook relay ativo
- preflight passando
- doctor passando

## Artefatos esperados

Gerados apenas localmente em `.e2e/webhook-ubyx/`:

- `e2e.log`
- `state.json`
- `mock-events.jsonl`
- `create_payment_request_broadcast.json`
- `create_payment_request_tx.json`
- `payment_request_query.json`
- `payment_request_query.http.txt`
- `payment_request_qr.json`
- `payment_request_qr.http.txt`
- `request_id.txt`
- `e2e_memo.txt`
- `pay_request_broadcast.json`
- `pay_request_tx.json`
- `pay_request_command.txt`

## Criterios de sucesso

- `create-merchant` aceito ou prerequisito ja satisfeito
- `create-payment-request` com `code=0`
- `REQUEST_ID` resolvido
- `GET /byx/payments/v1/payment_requests/{id}` retorna `200`
- `GET /byx/payments/v1/payment_requests/{id}/qr` retorna `200`
- `pay-payment-request` executado
- webhook entregue ao mock
- replay/idempotencia validado
- assinatura invalida rejeitada
- marcador final observado: `E2E_UBYX_OK request_id=<id>`

## Troubleshooting conhecido

### `tx not indexed`

- causa: consulta de tx antes da indexacao RPC
- acao: usar `wait_tx` com retry e timeout
- diagnostico manual:

```bash
./bin/byxd q tx <TXHASH> --node http://127.0.0.1:26657 -o json | jq
curl -s http://127.0.0.1:26657/status | jq -r '.result.node_info.other.tx_index'
```

### `signature verification failed`

- causa: tx rejeitada em `CheckTx`
- verificar:
  - `account_number`
  - `sequence`
  - `chain-id`
  - alinhamento entre key local e conta financiada

### `merchant not found`

- causa: chave `merchant` no keyring nao implica merchant registrado no modulo `lojas`
- acao: consultar/criar merchant antes do `create-payment-request`

### `invalid json` em broadcast

- causa: mistura de `gas estimate` ou stderr com stdout JSON
- acao: separar:
  - `.json`
  - `.raw.txt`
  - `.stderr.txt`

### Falha silenciosa apos `REQUEST_ID`

- causa historica: `curl -sf` e validacoes frageis com `set -e`
- acao: usar `curl_json_checked` e os marcadores:
  - `ASSERT_QUERY_FIELDS_START`
  - `ASSERT_QUERY_FIELDS_OK`
  - `PAY_REQUEST_START`
  - `PAY_REQUEST_DONE`

### `created` vs `reused`

- `CreatePaymentRequest` pode emitir:
  - `byx_payment_request_created`
  - `byx_payment_request_reused`
- a dedupe usa `loja_id + amount_ubyx + memo`
- o E2E usa memo unico por execucao para reduzir colisao entre rodadas

## Alerta de commit

- `.e2e/` e artefato local de diagnostico
- `.e2e/` nao deve ser commitado
- manter `.e2e/` no `.gitignore`
