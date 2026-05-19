#!/usr/bin/env bash
set -euo pipefail

# Local single-node runner for the evmd fork with JSON-RPC/WS enabled.
# It builds the binary (if needed), resets the home, seeds one dev key,
# patches app.toml for RPC/WS, and starts the node in the foreground.

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${BIN:-$ROOT/evmd-fork/build/evmd}"
HOME_DIR="${HOME_DIR:-$HOME/.evmd_localnet}"
CHAIN_ID="${CHAIN_ID:-byx_1}"
DENOM="${DENOM:-byx}"
STAKE_DENOM="${STAKE_DENOM:-byx}"
KEY_NAME="${KEY_NAME:-alice}"
KEYRING_BACKEND="${KEYRING_BACKEND:-test}"
RPC_ADDR="${RPC_ADDR:-127.0.0.1:8545}"
WS_ADDR="${WS_ADDR:-127.0.0.1:8546}"

echo "==> evmd-fork localnet"
echo "BIN=$BIN"
echo "HOME_DIR=$HOME_DIR"
echo "CHAIN_ID=$CHAIN_ID"
echo "DENOM=$DENOM"
echo "RPC=$RPC_ADDR WS=$WS_ADDR"

mkdir -p "$(dirname "$BIN")"

if [ ! -x "$BIN" ]; then
  echo "==> Building evmd binary..."
  (cd "$ROOT/evmd-fork" && GOTOOLCHAIN=local GOWORK=off go build -o "$BIN" ./cmd/evmd)
fi

echo "==> Reset home..."
rm -rf "$HOME_DIR"

echo "==> Init chain..."
"$BIN" init local --chain-id "$CHAIN_ID" --home "$HOME_DIR"

# force chain-id in genesis to avoid mismatch
jq ".chain_id = \"$CHAIN_ID\"" "$HOME_DIR/config/genesis.json" > "$HOME_DIR/config/genesis.json.tmp" \
  && mv "$HOME_DIR/config/genesis.json.tmp" "$HOME_DIR/config/genesis.json"

echo "==> Set client config chain-id..."
"$BIN" config set client chain-id "$CHAIN_ID" --home "$HOME_DIR" >/dev/null 2>&1 || true

echo "==> Set denom across genesis..."
jq --arg denom "$DENOM" '
  .app_state.evm.params.evm_denom = $denom
  | .app_state.evm.params.extended_denom_options.extended_denom = $denom
  | .app_state.mint.params.mint_denom = $denom
  | .app_state.staking.params.bond_denom = $denom
  | .app_state.gov.params.min_deposit = (.app_state.gov.params.min_deposit // [] | map(.denom = $denom))
  | .app_state.gov.params.expedited_min_deposit = (.app_state.gov.params.expedited_min_deposit // [] | map(.denom = $denom))
  | .app_state.feesplit.params.denoms_allowlist = [ $denom ]
  | .app_state.bank.denom_metadata = (
      (.app_state.bank.denom_metadata // [])
      | map(select(.base != $denom))
      + [
        {
          description: "BYX staking denom",
          denom_units: [
            {denom: $denom, exponent: 0, aliases: []},
            {denom: "BYX", exponent: 18, aliases: []}
          ],
          base: $denom,
          display: "BYX",
          name: "BYX",
          symbol: "BYX"
        }
      ]
    )
' "$HOME_DIR/config/genesis.json" > "$HOME_DIR/config/genesis.json.tmp" \
  && mv "$HOME_DIR/config/genesis.json.tmp" "$HOME_DIR/config/genesis.json"

echo "==> Create dev key ($KEY_NAME)..."
if ! "$BIN" keys show "$KEY_NAME" --home "$HOME_DIR" --keyring-backend "$KEYRING_BACKEND" >/dev/null 2>&1; then
  "$BIN" keys add "$KEY_NAME" --home "$HOME_DIR" --keyring-backend "$KEYRING_BACKEND" --output json
fi

echo "==> Fund genesis account..."
"$BIN" genesis add-genesis-account "$KEY_NAME" "1000000000000000000000${DENOM}" --home "$HOME_DIR" --keyring-backend "$KEYRING_BACKEND"

echo "==> Gentx..."
"$BIN" genesis gentx "$KEY_NAME" "500000000000000000000${STAKE_DENOM}" --home "$HOME_DIR" --keyring-backend "$KEYRING_BACKEND" --chain-id "$CHAIN_ID"
"$BIN" genesis collect-gentxs --home "$HOME_DIR"

APP_TOML="$HOME_DIR/config/app.toml"
if [ -f "$APP_TOML" ]; then
  echo "==> Patching app.toml for JSON-RPC/WS..."
  python3 - <<PY
from pathlib import Path
p = Path("$APP_TOML")
txt = p.read_text()
out = []
inside = False
for line in txt.splitlines():
    if line.strip() == "[json-rpc]":
        inside = True
        out.append(line)
        continue
    if inside and line.startswith("[") and line.strip().startswith("[") and not line.strip().startswith("[json-rpc]"):
        inside = False
    if inside and line.strip().startswith("enable"):
        out.append("enable = true")
    else:
        out.append(line)
p.write_text("\n".join(out) + ("\n" if txt.endswith("\n") else ""))
PY
  perl -0777 -pi -e 's/(\\[api\\][^\\[]*address\\s*=\\s*)\"[^\"]*\"/${1}\"127.0.0.1:1317\"/m' "$APP_TOML"
  perl -0777 -pi -e 's/(\\[grpc\\][^\\[]*address\\s*=\\s*)\"[^\"]*\"/${1}\"127.0.0.1:9090\"/m' "$APP_TOML"
  perl -pi -e "s#^address\\s*=\\s*\"[^\"]*\"#address = \"${RPC_ADDR}\"#" "$APP_TOML"
  perl -pi -e "s#^ws-address\\s*=\\s*\"[^\"]*\"#ws-address = \"${WS_ADDR}\"#" "$APP_TOML"
  # expand APIs if the line exists
  perl -pi -e 's/^api\\s*=\\s*\"[^\"]*\"/api = \"eth,net,web3,txpool,debug\"/' "$APP_TOML"
  perl -0777 -pi -e 's/(# EnableIndexer.*\\n)\\s*enable\\s*=\\s*.*$/\\1enable-indexer = false/m' "$APP_TOML"
  perl -0777 -pi -e 's/(Enabled profiling.*\\n)enable\\s*=\\s*.*$/\\1enable-profiling = false/m' "$APP_TOML"
fi

CONFIG_TOML="$HOME_DIR/config/config.toml"
if [ -f "$CONFIG_TOML" ]; then
  echo "==> Binding RPC to localhost..."
  perl -pi -e "s#^laddr\\s*=\\s*\"tcp://[^:]+:26657\"#laddr = \"tcp://127.0.0.1:26657\"#" "$CONFIG_TOML"
fi

echo "==> Starting node (Ctrl+C to stop)..."
exec "$BIN" start --home "$HOME_DIR" --keyring-backend "$KEYRING_BACKEND"
