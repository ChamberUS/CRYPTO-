# Sprint 11.5 — EVM compat matrix (EVM tag × Geth pin × Go toolchain)

Marca **OK** somente se: `go mod tidy` + `go build ./...` retornarem 0.

## Ambiente base
- go: go version go1.25.3 darwin/amd64
- cwd: /Users/buynnex-corp/dev/BYX\

## Matriz

| cosmos/evm | GETH_PIN | GOTOOLCHAIN | build | cosmos/evm resolved | go-ethereum resolved | log |
|---|---|---|---:|---|---|---|
| `v0.3.0` | `none` | `auto` | ❌ FAIL | `<not found>` | `<not found>` | `docs/evm_probe_v0.3.0_none_auto.log` |
| `v0.3.0` | `none` | `go1.22.13` | ❌ FAIL | `<not found>` | `<not found>` | `docs/evm_probe_v0.3.0_none_go1.22.13.log` |
| `v0.3.0` | `none` | `go1.23.6` | ❌ FAIL | `<not found>` | `<not found>` | `docs/evm_probe_v0.3.0_none_go1.23.6.log` |
| `v0.3.0` | `v1.14.13` | `auto` | ❌ FAIL | `<not found>` | `<not found>` | `docs/evm_probe_v0.3.0_v1.14.13_auto.log` |
| `v0.3.0` | `v1.14.13` | `go1.22.13` | ❌ FAIL | `<not found>` | `<not found>` | `docs/evm_probe_v0.3.0_v1.14.13_go1.22.13.log` |
| `v0.3.0` | `v1.14.13` | `go1.23.6` | ❌ FAIL | `<not found>` | `<not found>` | `docs/evm_probe_v0.3.0_v1.14.13_go1.23.6.log` |
| `v0.3.0` | `v1.14.11` | `auto` | ❌ FAIL | `<not found>` | `<not found>` | `docs/evm_probe_v0.3.0_v1.14.11_auto.log` |
| `v0.3.0` | `v1.14.11` | `go1.22.13` | ❌ FAIL | `<not found>` | `<not found>` | `docs/evm_probe_v0.3.0_v1.14.11_go1.22.13.log` |
| `v0.3.0` | `v1.14.11` | `go1.23.6` | ❌ FAIL | `<not found>` | `<not found>` | `docs/evm_probe_v0.3.0_v1.14.11_go1.23.6.log` |
| `v0.3.0` | `v1.13.15` | `auto` | ❌ FAIL | `<not found>` | `<not found>` | `docs/evm_probe_v0.3.0_v1.13.15_auto.log` |
| `v0.3.0` | `v1.13.15` | `go1.22.13` | ❌ FAIL | `<not found>` | `<not found>` | `docs/evm_probe_v0.3.0_v1.13.15_go1.22.13.log` |
| `v0.3.0` | `v1.13.15` | `go1.23.6` | ❌ FAIL | `<not found>` | `<not found>` | `docs/evm_probe_v0.3.0_v1.13.15_go1.23.6.log` |
| `v0.3.0` | `v1.12.2` | `auto` | ❌ FAIL | `<not found>` | `<not found>` | `docs/evm_probe_v0.3.0_v1.12.2_auto.log` |
| `v0.3.0` | `v1.12.2` | `go1.22.13` | ❌ FAIL | `<not found>` | `<not found>` | `docs/evm_probe_v0.3.0_v1.12.2_go1.22.13.log` |
| `v0.3.0` | `v1.12.2` | `go1.23.6` | ❌ FAIL | `<not found>` | `<not found>` | `docs/evm_probe_v0.3.0_v1.12.2_go1.23.6.log` |
| `v0.3.1` | `none` | `auto` | ❌ FAIL | `<not found>` | `<not found>` | `docs/evm_probe_v0.3.1_none_auto.log` |
| `v0.3.1` | `none` | `go1.22.13` | ❌ FAIL | `<not found>` | `<not found>` | `docs/evm_probe_v0.3.1_none_go1.22.13.log` |
| `v0.3.1` | `none` | `go1.23.6` | ❌ FAIL | `<not found>` | `<not found>` | `docs/evm_probe_v0.3.1_none_go1.23.6.log` |
| `v0.3.1` | `v1.14.13` | `auto` | ❌ FAIL | `<not found>` | `<not found>` | `docs/evm_probe_v0.3.1_v1.14.13_auto.log` |
