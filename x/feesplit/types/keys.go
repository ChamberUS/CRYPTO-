package types

import (
	"cosmossdk.io/collections"
	lojastypes "github.com/buynnex-corp/byx/x/lojas/types"
)

const (
	ModuleName = "feesplit"

	StoreKey = ModuleName

	RouterKey = ModuleName

	TreasuryModuleAccount = "treasury"
	FeeSplitModuleAccount = ModuleName

	DefaultDenom = lojastypes.BaseDenom
)

var (
	ParamsKey = collections.NewPrefix(0)
)
