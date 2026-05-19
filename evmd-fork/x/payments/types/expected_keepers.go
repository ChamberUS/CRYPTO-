package types

import (
	"context"

	lojastypes "github.com/buynnex/iaos-evmd/x/lojas/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected interface for the Bank module.
type BankKeeper interface {
	SendCoins(context.Context, sdk.AccAddress, sdk.AccAddress, sdk.Coins) error
}

// LojasKeeper defines the expected subset we need from x/lojas.
type LojasKeeper interface {
	GetMerchant(context.Context, uint64) (lojastypes.Merchant, error)
}
