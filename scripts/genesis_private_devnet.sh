#!/usr/bin/env bash
set -euo pipefail

CHAIN_ID="${CHAIN_ID:-byx-devnet-private-1}"
MONIKER="${MONIKER:-byx-devnet}"
HOME_DIR="${HOME_DIR:-$(pwd)/.byx-devnet-private}"
KEYRING_BACKEND="${KEYRING_BACKEND:-test}"
GENESIS_KEY_NAME="${GENESIS_KEY_NAME:-genesis}"
VALIDATOR_KEY_NAME="${VALIDATOR_KEY_NAME:-validator}"
FAUCET_ADMIN_KEY_NAME="${FAUCET_ADMIN_KEY_NAME:-faucet-admin}"
OUT_GENESIS="${OUT_GENESIS:-$(pwd)/genesis.json}"

TOTAL_SUPPLY_BYX="${TOTAL_SUPPLY_BYX:-1000000000}"
LOJAS_RESERVE_BYX="${LOJAS_RESERVE_BYX:-200000000}"
VALIDATOR_SELF_DELEGATION_BYX="${VALIDATOR_SELF_DELEGATION_BYX:-1000000}"
UBYX_PER_BYX=1000000

if ! command -v byxd >/dev/null 2>&1; then
  echo "byxd não encontrado no PATH" >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "jq não encontrado no PATH" >&2
  exit 1
fi

if [ "${LOJAS_RESERVE_BYX}" -ge "${TOTAL_SUPPLY_BYX}" ]; then
  echo "LOJAS_RESERVE_BYX deve ser menor que TOTAL_SUPPLY_BYX" >&2
  exit 1
fi

GENESIS_ACCOUNT_BYX=$((TOTAL_SUPPLY_BYX - LOJAS_RESERVE_BYX - VALIDATOR_SELF_DELEGATION_BYX))
if [ "${GENESIS_ACCOUNT_BYX}" -lt 0 ]; then
  echo "Distribuição inválida: TOTAL_SUPPLY_BYX insuficiente para reserve + self delegation" >&2
  exit 1
fi
if [ "${VALIDATOR_SELF_DELEGATION_BYX}" -gt "${TOTAL_SUPPLY_BYX}" ]; then
  echo "VALIDATOR_SELF_DELEGATION_BYX não pode exceder GENESIS_ACCOUNT_BYX" >&2
  exit 1
fi

TOTAL_SUPPLY_UBYX=$((TOTAL_SUPPLY_BYX * UBYX_PER_BYX))
LOJAS_RESERVE_UBYX=$((LOJAS_RESERVE_BYX * UBYX_PER_BYX))
VALIDATOR_SELF_DELEGATION_UBYX=$((VALIDATOR_SELF_DELEGATION_BYX * UBYX_PER_BYX))
GENESIS_ACCOUNT_UBYX=$((GENESIS_ACCOUNT_BYX * UBYX_PER_BYX))

echo "==> Limpando HOME_DIR: ${HOME_DIR}"
rm -rf "${HOME_DIR}"

echo "==> byxd init"
byxd init "${MONIKER}" --chain-id "${CHAIN_ID}" --home "${HOME_DIR}" >/dev/null 2>&1

echo "==> Criando chaves locais (keyring=${KEYRING_BACKEND})"
byxd keys add "${GENESIS_KEY_NAME}" --home "${HOME_DIR}" --keyring-backend "${KEYRING_BACKEND}" >/dev/null
byxd keys add "${VALIDATOR_KEY_NAME}" --home "${HOME_DIR}" --keyring-backend "${KEYRING_BACKEND}" >/dev/null
byxd keys add "${FAUCET_ADMIN_KEY_NAME}" --home "${HOME_DIR}" --keyring-backend "${KEYRING_BACKEND}" >/dev/null

GENESIS_ADDR="$(byxd keys show "${GENESIS_KEY_NAME}" -a --home "${HOME_DIR}" --keyring-backend "${KEYRING_BACKEND}")"
VALIDATOR_ADDR="$(byxd keys show "${VALIDATOR_KEY_NAME}" -a --home "${HOME_DIR}" --keyring-backend "${KEYRING_BACKEND}")"
FAUCET_ADMIN_ADDR="$(byxd keys show "${FAUCET_ADMIN_KEY_NAME}" -a --home "${HOME_DIR}" --keyring-backend "${KEYRING_BACKEND}")"
LOJAS_MODULE_ADDR="$(go run ./scripts/tools/module_address/main.go lojas)"

echo "==> Endereços"
echo "    genesis:       ${GENESIS_ADDR}"
echo "    validator:     ${VALIDATOR_ADDR}"
echo "    faucet-admin:  ${FAUCET_ADMIN_ADDR}"
echo "    lojas-module:  ${LOJAS_MODULE_ADDR}"

echo "==> Pré-funding genesis accounts"
byxd genesis add-genesis-account "${GENESIS_ADDR}" "${GENESIS_ACCOUNT_UBYX}ubyx" --home "${HOME_DIR}" --keyring-backend "${KEYRING_BACKEND}" >/dev/null
byxd genesis add-genesis-account "${VALIDATOR_ADDR}" "${VALIDATOR_SELF_DELEGATION_UBYX}ubyx" --home "${HOME_DIR}" --keyring-backend "${KEYRING_BACKEND}" >/dev/null
byxd genesis add-genesis-account "${LOJAS_MODULE_ADDR}" "${LOJAS_RESERVE_UBYX}ubyx" --home "${HOME_DIR}" >/dev/null

echo "==> Gentx + collect-gentxs"
byxd genesis gentx "${VALIDATOR_KEY_NAME}" "${VALIDATOR_SELF_DELEGATION_UBYX}ubyx" --chain-id "${CHAIN_ID}" --home "${HOME_DIR}" --keyring-backend "${KEYRING_BACKEND}" >/dev/null 2>&1
byxd genesis collect-gentxs --home "${HOME_DIR}" >/dev/null 2>&1

echo "==> Ajustando denom/tokenomics/faucet no genesis"
GENESIS_PATH="${HOME_DIR}/config/genesis.json"
jq \
  --arg chain_id "${CHAIN_ID}" \
  --arg faucet_admin "${FAUCET_ADMIN_ADDR}" \
  --argjson total_supply "${TOTAL_SUPPLY_UBYX}" \
  '
  .chain_id = $chain_id
  | .app_state.staking.params.bond_denom = "ubyx"
  | .app_state.gov.params.min_deposit = [{"denom":"ubyx","amount":"10000000000000"}]
  | .app_state.gov.params.expedited_min_deposit = [{"denom":"ubyx","amount":"50000000000000"}]
  | .app_state.mint.minter.inflation = "0.000000000000000000"
  | .app_state.mint.params.mint_denom = "ubyx"
  | .app_state.mint.params.inflation_rate_change = "0.000000000000000000"
  | .app_state.mint.params.inflation_max = "0.000000000000000000"
  | .app_state.mint.params.inflation_min = "0.000000000000000000"
  | .app_state.bank.denom_metadata = [
      {
        "description": "Token oficial da rede BYX",
        "denom_units": [
          {"denom":"ubyx","exponent":0},
          {"denom":"BYX","exponent":6}
        ],
        "base":"ubyx",
        "display":"BYX",
        "name":"BYX Token",
        "symbol":"BYX"
      }
    ]
  | .app_state.lojas.params.faucet_enabled = false
  | .app_state.lojas.params.faucet_admin = $faucet_admin
  | .app_state.bank.supply |= (map(if .denom == "ubyx" then .amount = ($total_supply|tostring) else . end))
  ' "${GENESIS_PATH}" > "${GENESIS_PATH}.tmp"
mv "${GENESIS_PATH}.tmp" "${GENESIS_PATH}"

echo "==> Validando genesis"
byxd genesis validate "${GENESIS_PATH}"

echo "==> Copiando genesis para ${OUT_GENESIS}"
cp "${GENESIS_PATH}" "${OUT_GENESIS}"

echo "OK: genesis privado pronto e validado."
