#!/usr/bin/env bash
set -euo pipefail

# Matrix probe for cosmos/evm compatibility.
# Dimensions: EVM_TAGS, GETH_PINS, TOOLCHAINS.
# Marks OK only if go mod tidy + go build ./... succeed.

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="$ROOT/docs/sprint11_5_evm_matrix_report.md"
mkdir -p "$ROOT/docs"

# Defaults
EVM_TAGS_DEFAULT=("v0.3.0" "v0.3.1" "v0.4.1" "v0.4.2" "v0.5.0")
GETH_PINS_DEFAULT=("none" "v1.14.13" "v1.14.11" "v1.13.15" "v1.12.2")
TOOLCHAINS_DEFAULT=("auto" "go1.22.13" "go1.23.6")

read_arr() {
  # $1=envvar name, $2=array name to fill, $3=defaults array (by name)
  local var="$1"
  local arrname="$2"
  local defname="$3"
  if [ -n "${!var:-}" ]; then
    # shellcheck disable=SC2206
    eval "$arrname=(${!var})"
  else
    eval "$arrname=(\"\${${defname}[@]}\")"
  fi
}

read_arr EVM_TAGS EVM_TAGS_ARR EVM_TAGS_DEFAULT
read_arr GETH_PINS GETH_PINS_ARR GETH_PINS_DEFAULT
read_arr TOOLCHAINS TOOLCHAINS_ARR TOOLCHAINS_DEFAULT

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

md(){ printf "%s\n" "$*" >> "$OUT"; }

modver() {
  local mod="$1"
  local v
  v="$(go list -m -f '{{if eq .Path "'"$mod"'"}}{{.Version}}{{end}}' all 2>/dev/null | tail -n 1 | tr -d '[:space:]' || true)"
  [ -n "$v" ] && echo "$v" || echo "<not found>"
}

write_header() {
  : > "$OUT"
  md "# Sprint 11.5 — EVM compat matrix (EVM tag × Geth pin × Go toolchain)"
  md ""
  md "Marca **OK** somente se: \`go mod tidy\` + \`go build ./...\` retornarem 0."
  md ""
  md "## Ambiente base"
  md "- go: $(go version)"
  md "- cwd: $ROOT"
  md "- date: $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
  md ""
  md "## Matriz"
  md ""
  md "| cosmos/evm | GETH_PIN | GOTOOLCHAIN | build | cosmos/evm resolved | go-ethereum resolved | log |"
  md "|---|---|---|---:|---|---|---|"
}

run_one() {
  local evm_tag="$1"
  local geth_pin="$2"
  local toolchain="$3"

  local work="$tmpdir/${evm_tag}_${geth_pin}_${toolchain}"
  mkdir -p "$work"

  cat > "$work/go.mod" <<MOD
module byx-evm-probe

go 1.22

require github.com/cosmos/evm ${evm_tag}
MOD

  cat > "$work/main.go" <<'GO'
package main
import _ "github.com/cosmos/evm/ethereum"
func main() {}
GO

  export GOWORK=off
  export GOPATH=
  if [ "$toolchain" != "auto" ]; then
    export GOTOOLCHAIN="$toolchain"
  else
    unset GOTOOLCHAIN || true
  fi

  if [ "$geth_pin" != "none" ]; then
    ( cd "$work" && go mod edit -replace "github.com/ethereum/go-ethereum=github.com/ethereum/go-ethereum@${geth_pin}" ) || true
  fi

  local safe_tag="${evm_tag//[^a-zA-Z0-9\.\-_]/_}"
  local safe_geth="${geth_pin//[^a-zA-Z0-9\.\-_]/_}"
  local safe_tc="${toolchain//[^a-zA-Z0-9\.\-_]/_}"
  local log="$ROOT/docs/evm_probe_${safe_tag}_${safe_geth}_${safe_tc}.log"
  : > "$log"

  set +e
  (
    cd "$work"
    echo "== go version ==" >>"$log"
    go version >>"$log" 2>&1
    echo "== go env ==" >>"$log"
    go env >>"$log" 2>&1
    echo "== go mod tidy ==" >>"$log"
    go mod tidy >>"$log" 2>&1
    echo "== go list -m all ==" >>"$log"
    go list -m all >>"$log" 2>&1
    echo "== go build ==" >>"$log"
    go build ./... >>"$log" 2>&1
  )
  rc=$?
  set -e

  local evm_resolved geth_resolved
  evm_resolved="$(cd "$work" && modver "github.com/cosmos/evm")"
  geth_resolved="$(cd "$work" && modver "github.com/ethereum/go-ethereum")"

  local status="FAIL"
  local emoji="❌"
  if [ "$rc" -eq 0 ]; then
    status="OK"
    emoji="✅"
  fi

  md "| \`${evm_tag}\` | \`${geth_pin}\` | \`${toolchain}\` | ${emoji} ${status} | \`${evm_resolved}\` | \`${geth_resolved}\` | \`docs/$(basename "$log")\` |"
}

write_footer() {
  md ""
  md "## Regras de decisão"
  md "- Se aparecer qualquer **OK**, escolher o combo mais moderno (evm mais novo) com toolchain estável (go1.22.x/1.23.x)."
  md "- Se só funcionar com evm v0.3.x, seguir temporariamente e planejar migração."
  md ""
  md "## Próximo passo (após OK)"
  md "1) Fixar cosmos/evm@<tag-ok> (e go-ethereum@<geth-ok> se preciso) no go.mod do BYX."
  md "2) Wiring atrás de EVM_ENABLED=1."
  md "3) RPC 8545 / WS 8546."
  md "4) MetaMask + deploy Solidity."
}

main() {
  write_header
  for evm in "${EVM_TAGS_ARR[@]}"; do
    for geth in "${GETH_PINS_ARR[@]}"; do
      for tc in "${TOOLCHAINS_ARR[@]}"; do
        echo "==> probing evm=$evm geth=$geth toolchain=$tc"
        run_one "$evm" "$geth" "$tc"
      done
    done
  done
  write_footer
  echo "==> Report: $OUT"
}

main "$@"
