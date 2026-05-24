#!/usr/bin/env bash
set -euo pipefail

# Smoke test for Sprint 12.1: x/payments MsgPayWithCertificate (atomic pay + certificate transfer).
# Requires a running node with REST enabled (1317) and CLI byxd in PATH.
#
# Defaults assume `ignite chain serve --reset-once` and default keys from config.yml (marcelo, roberta).

HOME_DIR="${HOME_DIR:-$HOME/.byx}"
CHAIN_ID="${CHAIN_ID:-byx}"
KEYRING_BACKEND="${KEYRING_BACKEND:-test}"
NODE="${NODE:-tcp://0.0.0.0:26657}"
REST="${REST:-http://127.0.0.1:1317}"

MERCHANT_KEY="${MERCHANT_KEY:-marcelo}"
PAYER_KEY="${PAYER_KEY:-roberta}"

MERCHANT_NAME="${MERCHANT_NAME:-IAOS Store}"
MERCHANT_ADDR="${MERCHANT_ADDR:-Rua 1}"
MERCHANT_OPERATOR="${MERCHANT_OPERATOR:-}"
MERCHANT_KYC_REF="${MERCHANT_KYC_REF:-smoke-kyc-ref}"
MERCHANT_DOCUMENT_HASH="${MERCHANT_DOCUMENT_HASH:-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa}"
MERCHANT_KYC_STATUS="${MERCHANT_KYC_STATUS:-approved}"

AMOUNT_UBYX="${AMOUNT_UBYX:-2000}"

CERT_CATEGORY="${CERT_CATEGORY:-NOTEBOOK}"
CERT_BRAND="${CERT_BRAND:-Dell}"
CERT_MODEL="${CERT_MODEL:-XPS}"
CERT_CONDITION="${CERT_CONDITION:-A}"
CERT_SEED="${CERT_SEED:-smoke-12-1}"

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
curl -fsS "${rpc_url}/status" >/dev/null
curl -fsS "${REST}/cosmos/base/tendermint/v1beta1/node_info" >/dev/null

merchant_addr="$(byxd keys show "${MERCHANT_KEY}" --keyring-backend "${KEYRING_BACKEND}" --home "${HOME_DIR}" -a)"
payer_addr="$(byxd keys show "${PAYER_KEY}" --keyring-backend "${KEYRING_BACKEND}" --home "${HOME_DIR}" -a)"

echo "merchant_owner: ${merchant_addr}"
echo "payer: ${payer_addr}"

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
    --from "${MERCHANT_KEY}" \
    --fees 20000ubyx \
    --chain-id "${CHAIN_ID}" \
    --node "${NODE}" \
    --home "${HOME_DIR}" \
    --broadcast-mode sync \
    --yes \
    --keyring-backend "${KEYRING_BACKEND}" \
    -o json)"

  txhash="$(echo "${tx_json}" | jq -r '.txhash // empty')"
  [[ -n "${txhash}" && "${txhash}" != "null" ]] || { echo "could not extract txhash"; exit 1; }

  for _ in $(seq 1 20); do
    sleep 1
    merchants_json="$(byxd query lojas merchant-all --node "${NODE}" -o json 2>/dev/null || true)"
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
  done

  echo "merchant creation did not reflect in queries in time" >&2
  exit 1
}

merchant_id="$(ensure_merchant "${merchant_addr}")"
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

echo "issuing certificate to payer..."
issue_tx="$(byxd tx certificados issue-certificate "${merchant_id}" "${CERT_CATEGORY}" "${CERT_BRAND}" "${CERT_MODEL}" "${serial_hash}" "${CERT_CONDITION}" "${image_uri}" "${image_sha256}" "${image_seed}" \
  --owner "${payer_addr}" \
  --from "${MERCHANT_KEY}" \
  --fees 20000ubyx \
  --chain-id "${CHAIN_ID}" \
  --node "${NODE}" \
  --home "${HOME_DIR}" \
  --broadcast-mode sync \
  --yes \
  --keyring-backend "${KEYRING_BACKEND}" \
  -o json)"
issue_txhash="$(echo "${issue_tx}" | jq -r '.txhash // empty')"
[[ -n "${issue_txhash}" && "${issue_txhash}" != "null" ]] || { echo "issue tx missing txhash"; exit 1; }

cert_id=""
for _ in $(seq 1 25); do
  sleep 1
  txq="$(byxd query tx "${issue_txhash}" --node "${NODE}" -o json 2>/dev/null || true)"
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
[[ -n "${cert_id}" && "${cert_id}" != "null" ]] || { echo "could not extract certificate id"; exit 1; }
echo "certificate_id: ${cert_id}"

echo "creating payment request..."
create_tx="$(byxd tx payments create-payment-request "${merchant_id}" "${AMOUNT_UBYX}" \
  --from "${MERCHANT_KEY}" \
  --fees 20000ubyx \
  --chain-id "${CHAIN_ID}" \
  --node "${NODE}" \
  --home "${HOME_DIR}" \
  --broadcast-mode sync \
  --yes \
  --keyring-backend "${KEYRING_BACKEND}" \
  -o json)"
create_txhash="$(echo "${create_tx}" | jq -r '.txhash // empty')"
[[ -n "${create_txhash}" && "${create_txhash}" != "null" ]] || { echo "create tx missing txhash"; exit 1; }

request_id=""
for _ in $(seq 1 25); do
  sleep 1
  txq="$(byxd query tx "${create_txhash}" --node "${NODE}" -o json 2>/dev/null || true)"
  request_id="$(echo "${txq}" | jq -r '
    (.events // [])
    | map(select(.type=="byx_payment_request_created"))
    | .[0].attributes // []
    | map(select(.key=="request_id"))
    | .[0].value // empty
  ' | head -n1)"
  if [[ -n "${request_id}" && "${request_id}" != "null" ]]; then
    break
  fi
done
[[ -n "${request_id}" && "${request_id}" != "null" ]] || { echo "could not extract request_id"; exit 1; }
echo "request_id: ${request_id}"

echo "querying QR payload via REST..."
qr_json="$(curl -fsS "${REST}/byx/payments/v1/payment_requests/${request_id}/qr")"
echo "${qr_json}" | jq -e --arg rid "${request_id}" --arg to "${merchant_addr}" '
  .request_id == $rid and .merchant_owner == $to and (.qr_payload | length) > 10
' >/dev/null

echo "paying with certificate (atomic)..."
pay_tx="$(byxd tx payments pay-with-certificate "${request_id}" "${cert_id}" \
  --from "${PAYER_KEY}" \
  --fees 20000ubyx \
  --chain-id "${CHAIN_ID}" \
  --node "${NODE}" \
  --home "${HOME_DIR}" \
  --broadcast-mode sync \
  --yes \
  --keyring-backend "${KEYRING_BACKEND}" \
  -o json)"
pay_txhash="$(echo "${pay_tx}" | jq -r '.txhash // empty')"
[[ -n "${pay_txhash}" && "${pay_txhash}" != "null" ]] || { echo "pay tx missing txhash"; exit 1; }

echo "validating state via CLI queries..."
pr_json="$(byxd query payments payment-request "${request_id}" --node "${NODE}" -o json)"
echo "${pr_json}" | jq -e --arg payer "${payer_addr}" '
  .payment_request.status == "PAYMENT_STATUS_PAID" and .payment_request.payer == $payer and (.payment_request.paid_at_unix|tonumber) > 0
' >/dev/null

cert_json="$(byxd query certificados certificate "${cert_id}" --node "${NODE}" -o json)"
echo "${cert_json}" | jq -e --arg owner "${merchant_addr}" '.certificate.owner == $owner' >/dev/null

echo "validating events in tx..."
txq="$(byxd query tx "${pay_txhash}" --node "${NODE}" -o json)"
echo "${txq}" | jq -e '
  ([.events[]? | .type] | index("byx_payment_request_paid")) != null and
  ([.events[]? | .type] | index("certificados_transfer")) != null and
  ([.events[]? | .type] | index("payments_paid_with_certificate")) != null
' >/dev/null

echo "OK: pay-with-certificate smoke passed"
