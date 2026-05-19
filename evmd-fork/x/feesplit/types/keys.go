package types

import "cosmossdk.io/collections"

const (
	ModuleName = "feesplit"

	StoreKey = ModuleName

	RouterKey = ModuleName

	TreasuryModuleAccount = "treasury"
	FeeSplitModuleAccount = ModuleName

	DefaultDenom = "byx"
)

var (
	ParamsKey = collections.NewPrefix(0)
)
