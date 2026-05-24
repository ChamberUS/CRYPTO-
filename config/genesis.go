package config

import (
	sdkmath "cosmossdk.io/math"
	lojastypes "github.com/buynnex-corp/byx/x/lojas/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TokenDenom = lojastypes.BaseDenom

// 1 bilhão de BYX em unidade base ubyx.
var InitialSupply = sdkmath.NewInt(lojastypes.MaxSupplyUbyx)

func GetGenesisSupply() sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin(TokenDenom, InitialSupply))
}
