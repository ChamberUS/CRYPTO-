#!/usr/bin/env bash
set -euo pipefail

# Smoke test for x/certificados (Sprint 12).
# Requires a running node with REST enabled (1317) and CLI byxd in PATH.

HOME_DIR="${HOME_DIR:-$HOME/.byx}"
CHAIN_ID="${CHAIN_ID:-byx}"
KEYRING_BACKEND="${KEYRING_BACKEND:-test}"
NODE="${NODE:-tcp://0.0.0.0:26657}"
REST="${REST:-http://127.0.0.1:1317}"

ISSUER_KEY="${ISSUER_KEY:-marcelo}"
NEW_OWNER_KEY="${NEW_OWNER_KEY:-roberta}"

MERCHANT_NAME="${MERCHANT_NAME:-IAOS Store}"
MERCHANT_ADDR="${MERCHANT_ADDR:-Rua 1}"
MERCHANT_OPERATOR="${MERCHANT_OPERATOR:-}"
MERCHANT_KYC_REF="${MERCHANT_KYC_REF:-smoke-kyc-ref}"
MERCHANT_DOCUMENT_HASH="${MERCHANT_DOCUMENT_HASH:-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa}"
MERCHANT_KYC_STATUS="${MERCHANT_KYC_STATUS:-approved}"

CERT_CATEGORY="${CERT_CATEGORY:-NOTEBOOK}"
CERT_BRAND="${CERT_BRAND:-Dell}"
CERT_MODEL="${CERT_MODEL:-XPS}"
CERT_CONDITION="${CERT_CONDITION:-A}"
CERT_SEED="${CERT_SEED:-smoke-seed-1}"

if ! command -v byxd >/dev/null 2>&1; then
  echo "byxd binary not found in PATH" >&2
  exit 1
fi
if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required for this script" >&2
  exit 1
fi
if ! command -v node >/dev/null 2>&1; then
  echo "node is required for image generation" >&2
  exit 1
fi

rpc_url="${NODE/tcp/http}"
if ! curl -fsS "${rpc_url}/status" >/dev/null 2>&1; then
  echo "❌ RPC (${NODE}) não respondeu." >&2
  exit 1
fi
if ! curl -fsS "${REST}/cosmos/base/tendermint/v1beta1/node_info" >/dev/null 2>&1; then
  echo "❌ REST (${REST}) não respondeu." >&2
  exit 1
fi

issuer_addr="$(byxd keys show "${ISSUER_KEY}" --keyring-backend "${KEYRING_BACKEND}" --home "${HOME_DIR}" -a)"
new_owner_addr="$(byxd keys show "${NEW_OWNER_KEY}" --keyring-backend "${KEYRING_BACKEND}" --home "${HOME_DIR}" -a)"

echo "issuer: ${issuer_addr}"
echo "new_owner: ${new_owner_addr}"

ensure_merchant() {
  local creator="$1"

  local merchants_json
  merchants_json="$(byxd query lojas merchant-all --node "${NODE}" -o json 2>/dev/null || true)"
  if [[ -n "${merchants_json}" ]]; then
    local existing_id
    existing_id="$(
      echo "${merchants_json}" | jq -r --arg creator "${creator}" --arg name "${MERCHANT_NAME}" '
        (.merchant // []) | map(select(.creator==$creator and .nome==$name)) | (max_by(.id) // empty) | .id // empty
      ' | head -n1
    )"
    if [[ -n "${existing_id}" && "${existing_id}" != "null" ]]; then
      echo "${existing_id}"
      return 0
    fi
  fi

  echo "merchant not found for creator; creating one..." >&2
  tx_json="$(byxd tx lojas create-merchant "${MERCHANT_NAME}" "${MERCHANT_ADDR}" "${MERCHANT_OPERATOR}" "${MERCHANT_KYC_REF}" "${MERCHANT_DOCUMENT_HASH}" "${MERCHANT_KYC_STATUS}" \
    --from "${ISSUER_KEY}" \
    --fees 20000ubyx \
    --chain-id "${CHAIN_ID}" \
    --node "${NODE}" \
    --home "${HOME_DIR}" \
    --broadcast-mode sync \
    --yes \
    --keyring-backend "${KEYRING_BACKEND}" \
    -o json)"

  txhash="$(echo "${tx_json}" | jq -r '.txhash // empty')"
  if [[ -z "${txhash}" || "${txhash}" == "null" ]]; then
    echo "could not extract txhash from create-merchant tx" >&2
    echo "${tx_json}" >&2
    exit 1
  fi

  for _ in $(seq 1 20); do
    sleep 1
    merchants_json="$(byxd query lojas merchant-all --node "${NODE}" -o json 2>/dev/null || true)"
    if [[ -n "${merchants_json}" ]]; then
      local new_id
      new_id="$(
        echo "${merchants_json}" | jq -r --arg creator "${creator}" --arg name "${MERCHANT_NAME}" '
          (.merchant // []) | map(select(.creator==$creator and .nome==$name)) | (max_by(.id) // empty) | .id // empty
        ' | head -n1
      )"
      if [[ -n "${new_id}" && "${new_id}" != "null" ]]; then
        echo "${new_id}"
        return 0
      fi
    fi
  done

  echo "merchant creation did not reflect in queries in time" >&2
  exit 1
}

merchant_id="$(ensure_merchant "${issuer_addr}")"
echo "merchant_id: ${merchant_id}"

serial_number="SN-${CERT_BRAND}-${CERT_MODEL}-$(date +%s)"
serial_hash="$(printf '%s' "${serial_number}" | shasum -a 256 | awk '{print $1}')"

gen_json="$(node tools/cdp_image_gen/generate.js --json "$(jq -nc \
  --arg category "${CERT_CATEGORY}" \
  --arg brand "${CERT_BRAND}" \
  --arg model "${CERT_MODEL}" \
  --arg serial_hash "${serial_hash}" \
  --arg seed "${CERT_SEED}" \
  '{category:$category,brand:$brand,model:$model,serial_hash:$serial_hash,seed:$seed}')" )"

image_uri="$(echo "${gen_json}" | jq -r '.image_uri')"
image_sha256="$(echo "${gen_json}" | jq -r '.image_sha256')"
image_seed="$(echo "${gen_json}" | jq -r '.image_seed')"

echo "image_uri: ${image_uri}"
echo "image_sha256: ${image_sha256}"

echo "issuing certificate..."
issue_tx="$(byxd tx certificados issue-certificate "${merchant_id}" "${CERT_CATEGORY}" "${CERT_BRAND}" "${CERT_MODEL}" "${serial_hash}" "${CERT_CONDITION}" "${image_uri}" "${image_sha256}" "${image_seed}" \
  --from "${ISSUER_KEY}" \
  --fees 20000ubyx \
  --chain-id "${CHAIN_ID}" \
  --node "${NODE}" \
  --home "${HOME_DIR}" \
  --broadcast-mode sync \
  --yes \
  --keyring-backend "${KEYRING_BACKEND}" \
  -o json)"

txhash="$(echo "${issue_tx}" | jq -r '.txhash // empty')"
if [[ -z "${txhash}" || "${txhash}" == "null" ]]; then
  echo "could not extract txhash from issue-certificate tx" >&2
  echo "${issue_tx}" >&2
  exit 1
fi

cert_id=""
for _ in $(seq 1 25); do
  sleep 1
  txq="$(byxd query tx "${txhash}" --node "${NODE}" -o json 2>/dev/null || true)"
  cert_id="$(echo "${txq}" | jq -r '
    (.events // [])
    | map(select(.type=="certificados_issue"))
    | .[0].attributes // []
    | map(select(.key=="id"))
    | .[0].value // empty
  ' | head -n1)"
  if [[ -n "${cert_id}" && "${cert_id}" != "null" ]]; then
    break
  fi
done

if [[ -z "${cert_id}" || "${cert_id}" == "null" ]]; then
  echo "could not find certificate id from tx events" >&2
  byxd query tx "${txhash}" --node "${NODE}" -o json || true
  exit 1
fi

echo "certificate_id: ${cert_id}"

echo "querying via REST..."
cert_json="$(curl -fsS "${REST}/byx/certificados/v1/certificates/${cert_id}")"
echo "${cert_json}" | jq -e --arg id "${cert_id}" --arg owner "${issuer_addr}" --arg m "${merchant_id}" --arg sh "${serial_hash}" '
  .certificate.id == ($id|tonumber) and
  .certificate.owner == $owner and
  .certificate.merchant_id == ($m|tonumber) and
  (.certificate.serial_hash | ascii_downcase) == ($sh|ascii_downcase) and
  (.certificate.image_sha256 | length) >= 32
' >/dev/null

echo "transferring..."
transfer_tx="$(byxd tx certificados transfer-certificate "${cert_id}" "${new_owner_addr}" \
  --from "${ISSUER_KEY}" \
  --fees 20000ubyx \
  --chain-id "${CHAIN_ID}" \
  --node "${NODE}" \
  --home "${HOME_DIR}" \
  --broadcast-mode sync \
  --yes \
  --keyring-backend "${KEYRING_BACKEND}" \
  -o json)"
echo "${transfer_tx}" | jq -e '.txhash' >/dev/null

echo "re-querying via REST..."
cert_json2="$(curl -fsS "${REST}/byx/certificados/v1/certificates/${cert_id}")"
echo "${cert_json2}" | jq -e --arg owner "${new_owner_addr}" '.certificate.owner == $owner' >/dev/null

echo "OK: certificados smoke passed"
