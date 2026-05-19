# Sprint 11.3 — EVM Compatibility Report

Este arquivo é gerado/atualizado pelo script `scripts/evm_compat_probe.sh`.

Resultados do probe:

| Stack | Module@version | Status | Notas |
|-------|----------------|--------|-------|
| cosmos-evm | `github.com/cosmos/evm` | fail | # github.com/bytedance/sonic/internal/rt
/Users/buynnex-corp/go/pkg/mod/github.com/bytedance/sonic@v1.13.2/internal/rt/stubs.go:33:22: undefined: GoMapIterator
/Users/buynnex-corp/go/pkg/mod/github.com/bytedance/sonic@v1.13.2/internal/rt/stubs.go:36:54: undefined: GoMapIterator
FAIL	cosmos-evm-probe [build failed]
FAIL |
| ethermint | `github.com/evmos/ethermint` | fail | # ethermint-probe
imports_probe_test.go:5:3: no required module provides package github.com/evmos/ethermint; to add it:
	go get github.com/evmos/ethermint
FAIL	ethermint-probe [setup failed]
FAIL |
| evmos | `github.com/evmos/evmos/v17` | fail | # evmos-probe
imports_probe_test.go:5:3: no required module provides package github.com/evmos/evmos/v17; to add it:
	go get github.com/evmos/evmos/v17
FAIL	evmos-probe [setup failed]
FAIL |

Go version: go version go1.25.3 darwin/amd64
Probe done at: 2025-12-15T05:58:53-03:00
