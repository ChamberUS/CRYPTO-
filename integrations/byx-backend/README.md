# BYX Backend API

API minima para teste fechado do app BYX. Este servico e **somente DEVNET / TESTE FECHADO**.
Nao use esta API em mainnet, ambiente publico ou fluxo financeiro real.

O app mobile/web deve chamar esta API, nao o RPC `26657` diretamente. O backend acessa a chain via `byxd` local usando `BYX_NODE_RPC`, `CHAIN_ID`, `BYXD_HOME` e `KEYRING_BACKEND`.
Unidade monetaria: base on-chain `ubyx`, display `BYX`, com `1 BYX = 1_000_000 ubyx`.
Compatibilidade temporaria: se `BYX_NODE_RPC` nao existir, ele tenta `BYXD_NODE` e por ultimo `NODE`, mas rejeita qualquer valor que nao comece com `tcp://`, `http://` ou `https://`.

## Seguranca

- Nao coloque mnemonic, private key ou seed em variaveis de ambiente.
- `BYX_CREATE_PAYMENT_KEY` e `BYX_DEVNET_PAYER_KEY` sao nomes de chaves ja existentes no keyring local.
- Use `Authorization: Bearer $BYX_BACKEND_API_TOKEN` em todos os endpoints exceto health.
- Restrinja esta API por firewall/VPN/reverse proxy com auth; ela contem endpoint de pagamento devnet.
- Mantenha `26657` privado para localhost/VPN.
- Se `NODE_ENV=production`, o processo se recusa a subir sem `BYX_BACKEND_ALLOW_PRODUCTION=true`.

## Endpoints

- `GET /v1/devnet/health`
- `GET /v1/devnet/merchants/:id`
- `GET /v1/devnet/merchants/:id/saldo`
- `POST /v1/devnet/payment-requests`
- `GET /v1/devnet/payment-requests/:id`
- `GET /v1/devnet/payment-requests/:id/qr`
- `POST /v1/devnet/payment-requests/:id/pay`
- `GET /v1/devnet/game/petz/balance`
- `POST /v1/devnet/game/petz/reward`

Os endpoints `game/petz/*` sao placeholders para teste de integracao do app. `reward` exige Bearer auth mesmo se `BYX_ALLOW_UNAUTHENTICATED=true` e retorna `501` enquanto a regra de recompensa nao existir.

## Payloads

Criar payment request:

```json
{
  "loja_id": 1,
  "amount_ubyx": 2000,
  "memo": "pedido fechado",
  "expires_in_seconds": 300
}
```

`amount_ubyx` representa a unidade minima on-chain `ubyx`.

Pagar payment request em devnet:

```json
{}
```

O pagamento usa a chave local indicada em `BYX_DEVNET_PAYER_KEY`.

## Variaveis

Copie `.env.example` para `/etc/byx/byx-backend.env` ou carregue no ambiente do processo.

```sh
BYX_NODE_RPC=tcp://127.0.0.1:26657
CHAIN_ID=byx-devnet
BYXD_HOME=/var/lib/byxd
KEYRING_BACKEND=test
NODE_ENV=development
BYX_BACKEND_ALLOW_PRODUCTION=false
BYX_BACKEND_API_TOKEN=replace-with-long-random-devnet-token
BYX_CREATE_PAYMENT_KEY=merchant-key-name
BYX_DEVNET_PAYER_KEY=payer-key-name
```

## Rodar local

```sh
cd integrations/byx-backend
npm install
npm start
```

## systemd na VPS Locaweb

Premissas para a VPS Locaweb:

- `byxd` instalado em `/usr/local/bin/byxd`.
- Node.js 20 instalado.
- Chain rodando na mesma VPS com RPC apenas local em `tcp://127.0.0.1:26657`.
- Home da chain em `/var/lib/byxd` ou ajuste `BYXD_HOME`.
- Chaves ja importadas no keyring local por nome; nunca coloque mnemonic no `.env`.

Instalacao sugerida:

```sh
sudo mkdir -p /opt/byx
sudo rsync -a --delete ./ /opt/byx/

cd /opt/byx/integrations/byx-backend
npm install --omit=dev

sudo mkdir -p /etc/byx
sudo cp .env.example /etc/byx/byx-backend.env
sudo sed -i 's|BYXD_HOME=/var/lib/byxd|BYXD_HOME=/var/lib/byxd|' /etc/byx/byx-backend.env
sudo sed -i 's|NODE_ENV=development|NODE_ENV=production|' /etc/byx/byx-backend.env
sudo sed -i 's|BYX_BACKEND_ALLOW_PRODUCTION=false|BYX_BACKEND_ALLOW_PRODUCTION=true|' /etc/byx/byx-backend.env
sudo cp deploy/byx-backend.service /etc/systemd/system/byx-backend.service
sudo systemctl daemon-reload
sudo systemctl enable --now byx-backend
sudo systemctl status byx-backend --no-pager
```

Depois, coloque Nginx/Caddy na frente com TLS e restrinja acesso por IP/VPN ou auth adicional. Nao exponha `26657/tcp` publicamente.

## Docker

O container precisa ter o binario `byxd` disponivel e acesso ao `BYXD_HOME` com keyring local.

```sh
docker build -t byx-backend:devnet .
docker run --rm -p 8080:8080 \
  --env-file .env.example \
  -e BYX_BACKEND_ALLOW_PRODUCTION=true \
  -v /var/lib/byxd:/var/lib/byxd \
  -v /usr/local/bin/byxd:/usr/local/bin/byxd:ro \
  byx-backend:devnet
```

## Logs

As transacoes registram JSON em stdout/stderr com `txhash`, `request_id`, `loja_id` e `status`.
Por padrao, o backend aguarda ate `BYX_TX_WAIT_MS` pelo recibo da tx para extrair eventos como `request_id`.
