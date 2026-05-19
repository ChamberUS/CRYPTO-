# Sprint 11.4 — EVM Compat Probe (Cosmos EVM v0.5.0)

Probe: cria um mini módulo Go isolado e tenta compilar contra github.com/cosmos/evm.

## Ambiente
- go: go version go1.25.3 darwin/amd64
- cwd: /Users/buynnex-corp/dev/BYX

## Teste A: cosmos/evm @ v0.5.0
- ❌ build falhou

### Log (trecho)
```
go: downloading github.com/cosmos/evm v0.5.0
go: downloading github.com/cosmos/cosmos-sdk v0.53.4
go: downloading github.com/cometbft/cometbft v0.38.18
go: downloading google.golang.org/protobuf v1.36.8
go: downloading github.com/spf13/cast v1.9.2
go: downloading github.com/bytedance/sonic v1.14.0
go: downloading google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7
go: downloading github.com/spf13/pflag v1.0.9
go: downloading cosmossdk.io/client/v2 v2.0.0-beta.7
go: downloading github.com/onsi/gomega v1.38.0
go: downloading github.com/btcsuite/btcd v0.24.2
go: downloading github.com/tyler-smith/go-bip39 v1.1.0
go: downloading github.com/btcsuite/btcd/btcec/v2 v2.3.4
go: downloading github.com/ferranbt/fastssz v0.1.4
go: downloading github.com/btcsuite/btcd/chaincfg/chainhash v1.1.0
go: downloading github.com/minio/sha256-simd v1.0.0
go: downloading github.com/mitchellh/mapstructure v1.5.0
go: downloading go.yaml.in/yaml/v3 v3.0.3
go: downloading github.com/onsi/ginkgo v1.16.5
go: downloading github.com/crate-crypto/go-kzg-4844 v1.1.0
go: downloading github.com/leanovate/gopter v0.2.11
go: downloading github.com/cespare/cp v0.1.0
go: downloading github.com/allegro/bigcache v1.2.1-0.20190218064605-e24eb225f156
go: downloading github.com/gballet/go-libpcsclite v0.0.0-20190607065134-2772fd86a8ff
go: downloading github.com/urfave/cli/v2 v2.27.5
go: downloading github.com/golang-jwt/jwt/v4 v4.5.1
go: downloading github.com/peterh/liner v1.1.1-0.20190123174540-a2c9a5303de7
go: downloading github.com/holiman/billy v0.0.0-20240216141850-2abb0c79d3c4
go: downloading github.com/graph-gophers/graphql-go v1.3.0
go: downloading github.com/hashicorp/go-bexpr v0.1.10
go: downloading gopkg.in/natefinch/lumberjack.v2 v2.2.1
go: downloading github.com/influxdata/influxdb-client-go/v2 v2.4.0
go: downloading github.com/influxdata/influxdb1-client v0.0.0-20220302092344-a9ab5670611c
go: downloading github.com/opencontainers/image-spec v1.1.0-rc5
go: downloading github.com/docker/go-connections v0.5.0
go: downloading github.com/urfave/cli v1.22.1
go: downloading github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161
go: downloading github.com/mitchellh/pointerstructure v1.2.0
go: downloading github.com/opentracing/opentracing-go v1.2.0
go: downloading github.com/influxdata/line-protocol v0.0.0-20200327222509-2487e7298839
go: downloading github.com/deepmap/oapi-codegen v1.6.0
go: downloading github.com/cpuguy83/go-md2man/v2 v2.0.6
go: downloading github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1
# github.com/cosmos/evm/eips
/Users/buynnex-corp/go/pkg/mod/github.com/cosmos/evm@v0.5.0/eips/eips.go:15:36: jt[vm.CREATE].GetConstantGas undefined (type *vm.operation has no field or method GetConstantGas)
/Users/buynnex-corp/go/pkg/mod/github.com/cosmos/evm@v0.5.0/eips/eips.go:16:16: jt[vm.CREATE].SetConstantGas undefined (type *vm.operation has no field or method SetConstantGas)
/Users/buynnex-corp/go/pkg/mod/github.com/cosmos/evm@v0.5.0/eips/eips.go:18:38: jt[vm.CREATE2].GetConstantGas undefined (type *vm.operation has no field or method GetConstantGas)
/Users/buynnex-corp/go/pkg/mod/github.com/cosmos/evm@v0.5.0/eips/eips.go:19:17: jt[vm.CREATE2].SetConstantGas undefined (type *vm.operation has no field or method SetConstantGas)
/Users/buynnex-corp/go/pkg/mod/github.com/cosmos/evm@v0.5.0/eips/eips.go:25:28: jt[vm.CALL].GetConstantGas undefined (type *vm.operation has no field or method GetConstantGas)
/Users/buynnex-corp/go/pkg/mod/github.com/cosmos/evm@v0.5.0/eips/eips.go:26:14: jt[vm.CALL].SetConstantGas undefined (type *vm.operation has no field or method SetConstantGas)
/Users/buynnex-corp/go/pkg/mod/github.com/cosmos/evm@v0.5.0/eips/eips.go:32:16: jt[vm.SSTORE].SetConstantGas undefined (type *vm.operation has no field or method SetConstantGas)
```

## Próximas ações sugeridas
- Se falhar por sonic/runtime/linkname: testar com Go 1.22.x, e/ou pin/replace do sonic.
- Se build ok: partir para wiring no app (feature-flag) + JSON-RPC :8545.
