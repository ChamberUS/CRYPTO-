package types

import (
	"context"

	lojastypes "github.com/buynnex-corp/byx/x/lojas/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected interface for the Bank module.
type BankKeeper interface {
	SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
}

// LojasKeeper defines the expected subset we need from x/lojas.
type LojasKeeper interface {
	GetMerchant(context.Context, uint64) (lojastypes.Merchant, error)
}
