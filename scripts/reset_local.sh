#!/usr/bin/env bash
echo "♻ Reset local BYX"

rm -rf ~/.byx
rm -f webhook-relay/state.json

echo "✔ ~/.byx removido"
echo "✔ state.json limpo"
echo "Agora reinicie a chain e recrie merchant"
