# BYX Backend API

API minima para teste fechado do app BYX. Este servico e **somente DEVNET / TESTE FECHADO**.

O app mobile/web deve chamar esta API, nao o RPC `26657` diretamente. O backend acessa a chain via `byxd` local usando `NODE`, `CHAIN_ID`, `BYXD_HOME` e `KEYRING_BACKEND`.

## Seguranca

- Nao coloque mnemonic, private key ou seed em variaveis de ambiente.
- `BYX_CREATE_PAYMENT_KEY` e `BYX_DEVNET_PAYER_KEY` sao nomes de chaves ja existentes no keyring local.
- Use `Authorization: Bearer $BYX_BACKEND_API_TOKEN` em todos os endpoints exceto health.
- Restrinja esta API por firewall/VPN/reverse proxy com auth; ela contem endpoint de pagamento devnet.
- Mantenha `26657` privado para localhost/VPN.

## Endpoints

- `GET /v1/devnet/health`
- `GET /v1/devnet/merchants/:id`
- `GET /v1/devnet/merchants/:id/saldo`
- `POST /v1/devnet/payment-requests`
- `GET /v1/devnet/payment-requests/:id`
- `GET /v1/devnet/payment-requests/:id/qr`
- `POST /v1/devnet/payment-requests/:id/pay`

## Payloads

Criar payment request:

```json
{
  "loja_id": 1,
  "amount_microbyx": 2000,
  "memo": "pedido fechado",
  "expires_in_seconds": 300
}
```

Pagar payment request em devnet:

```json
{}
```

O pagamento usa a chave local indicada em `BYX_DEVNET_PAYER_KEY`.

## Variaveis

Copie `.env.example` para `/etc/byx/byx-backend.env` ou carregue no ambiente do processo.

```sh
NODE=tcp://127.0.0.1:26657
CHAIN_ID=byx-devnet
BYXD_HOME=/var/lib/byxd
KEYRING_BACKEND=test
BYX_BACKEND_API_TOKEN=change-me-devnet-token
BYX_CREATE_PAYMENT_KEY=merchant-dev
BYX_DEVNET_PAYER_KEY=payer-dev
```

## Rodar local

```sh
cd integrations/byx-backend
npm install
npm start
```

## systemd

```sh
sudo mkdir -p /etc/byx
sudo cp .env.example /etc/byx/byx-backend.env
sudo cp deploy/byx-backend.service /etc/systemd/system/byx-backend.service
sudo systemctl daemon-reload
sudo systemctl enable --now byx-backend
```

## Docker

O container precisa ter o binario `byxd` disponivel e acesso ao `BYXD_HOME` com keyring local.

```sh
docker build -t byx-backend:devnet .
docker run --rm -p 8080:8080 \
  --env-file .env.example \
  -v /var/lib/byxd:/var/lib/byxd \
  -v /usr/local/bin/byxd:/usr/local/bin/byxd:ro \
  byx-backend:devnet
```

## Logs

As transacoes registram JSON em stdout/stderr com `txhash`, `request_id`, `loja_id` e `status`.
Por padrao, o backend aguarda ate `BYX_TX_WAIT_MS` pelo recibo da tx para extrair eventos como `request_id`.
