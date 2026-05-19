package config

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TokenDenom = "byx"

// 1 bilhão
var InitialSupply = sdkmath.NewInt(1_000_000_000)

func GetGenesisSupply() sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin(TokenDenom, InitialSupply))
}
