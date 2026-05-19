# BYX — Checklist AWS Devnet Privada

Objetivo: operar devnet privada na AWS sem exposição pública de RPC/API sensível.

## 1) Topologia mínima
- `validator` privado (sem RPC/API público).
- `sentry` público apenas para P2P.
- `rpc` (fullnode) separado para consultas internas/equipe.

## 2) Portas e Security Groups
- `22/tcp`: somente seu IP fixo.
- `26656/tcp`: P2P entre sentry/validator/peers autorizados.
- `26657/tcp`: apenas localhost/VPN/rede privada.
- `1317/tcp`: apenas localhost/VPN/rede privada.
- `9090/tcp`: apenas localhost/VPN/rede privada.
- `26660/tcp`: apenas rede interna (metrics).

## 3) Configuração do nó
- `minimum-gas-prices = "0.025byx"` no `app.toml`.
- `prometheus = true` somente para rede interna.
- `pex = false` no validator privado (recomendado).
- `persistent_peers` fixo entre validator/sentry.
- nunca expor `priv_validator_key.json`.

## 4) Genesis e tokenomics
- gerar genesis real com `scripts/genesis_private_devnet.sh`.
- validar com `byxd genesis validate genesis.json`.
- supply total `byx` fixada em `1_000_000_000`.
- reserva do módulo `lojas` pré-fundida no genesis.
- `x/mint` com inflação zerada no genesis.

## 5) systemd (produção privada)
Exemplo de unit:
```ini
[Unit]
Description=BYX Node
After=network-online.target

[Service]
User=byx
ExecStart=/usr/local/bin/byxd start --home /opt/byx --minimum-gas-prices 0.025byx
Restart=always
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

## 6) Backup e snapshots
- snapshot diário de disco (EBS snapshot).
- backup de `/config` e `/data` em bucket privado.
- testar restauração semanalmente.

## 7) Healthchecks operacionais
- status RPC: `curl -fsS http://127.0.0.1:26657/status`
- node info REST: `curl -fsS http://127.0.0.1:1317/cosmos/base/tendermint/v1beta1/node_info`
- bloco atual: `curl -fsS http://127.0.0.1:26657/status | jq -r '.result.sync_info.latest_block_height'`
- peers: `curl -fsS http://127.0.0.1:26657/net_info | jq -r '.result.n_peers'`

## 8) Regras de segurança
- `KEYRING_BACKEND=test` proibido em ambiente público.
- web-faucet apenas dev local; não expor internet.
- rotação periódica de chaves operacionais.
- branch protection + CI obrigatória antes de merge.

