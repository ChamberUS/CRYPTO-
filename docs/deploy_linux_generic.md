# BYX Deploy — Linux genérico (Ubuntu/Debian)

Este fluxo funciona em qualquer VPS root (Locaweb, EC2, Hetzner, etc.), sem dependência de AWS.

## Arquivos
- `scripts/deploy/.env.example`
- `scripts/deploy/bootstrap.sh`
- `scripts/deploy/configure_node.sh`
- `scripts/deploy/install_systemd.sh`
- `scripts/deploy/healthcheck.sh`
- `scripts/deploy/backup_snapshot.sh`

## 1) Preparar ambiente
```bash
sudo apt-get update
sudo apt-get install -y jq curl perl tar coreutils
```

Instale o `byxd` em `/usr/local/bin/byxd` (ou ajuste `BYXD_BIN` no `.env`).

## 2) Configurar variáveis
```bash
cp scripts/deploy/.env.example scripts/deploy/.env
```

Edite `scripts/deploy/.env`:
- `BYX_HOME`
- `BYX_CHAIN_ID`
- `BYX_GENESIS_FILE`
- `BYX_P2P_LADDR`
- `BYX_EXTERNAL_ADDRESS`
- `BYX_PERSISTENT_PEERS`
- `BYX_SEEDS`

Padrão seguro: RPC/API em localhost.

## 3) Bootstrap + hardening
```bash
sudo ENV_FILE=$(pwd)/scripts/deploy/.env bash scripts/deploy/bootstrap.sh
sudo ENV_FILE=$(pwd)/scripts/deploy/.env bash scripts/deploy/configure_node.sh
```

Dry-run sem gravar:
```bash
sudo ENV_FILE=$(pwd)/scripts/deploy/.env DRY_RUN=true bash scripts/deploy/configure_node.sh
```

Validação explícita (TOML + genesis):
```bash
sudo ENV_FILE=$(pwd)/scripts/deploy/.env bash scripts/deploy/validate_node_config.sh
```

## 4) Instalar systemd
```bash
sudo ENV_FILE=$(pwd)/scripts/deploy/.env bash scripts/deploy/install_systemd.sh
sudo systemctl start byxd
sudo systemctl status byxd --no-pager
```

## 5) Healthcheck
```bash
ENV_FILE=$(pwd)/scripts/deploy/.env bash scripts/deploy/healthcheck.sh
```

## 6) Backup/Snapshot genérico
```bash
sudo ENV_FILE=$(pwd)/scripts/deploy/.env bash scripts/deploy/backup_snapshot.sh
```

Opcional: agendar via cron/systemd timer.

## Portas recomendadas (firewall)
- `26656/tcp`: P2P (pública somente se necessário).
- `26657/tcp`: manter privado (localhost/VPN).
- `1317/tcp`: manter privado (localhost/VPN).
- `9090/tcp`: manter privado (localhost/VPN).
- `26660/tcp`: métricas privadas.
