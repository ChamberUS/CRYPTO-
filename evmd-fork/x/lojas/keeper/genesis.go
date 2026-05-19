// x/lojas/keeper/genesis.go
package keeper

import (
	"context"

	"github.com/buynnex/iaos-evmd/x/lojas/types"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis inicializa o estado do módulo a partir do genesis.
func (k Keeper) InitGenesis(ctx context.Context, genState types.GenesisState) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// 1) Params: se vier vazio, usa defaults; em seguida grava.
	params := genState.Params
	if (params == types.Params{}) {
		params = types.DefaultParams()
	}
	if err := k.ParamsStore.Set(sdkCtx, params); err != nil {
		return err
	}

	// 2) Merchants
	var maxID uint64
	for _, elem := range genState.MerchantList {
		if err := k.SetMerchant(sdkCtx, elem); err != nil {
			return err
		}
		if elem.Id > maxID {
			maxID = elem.Id
		}
		if elem.Creator != "" {
			k.SetMerchantByCreator(sdkCtx, elem.Creator, elem.Id)
		}
	}

	// 3) Sequência de IDs
	nextID := genState.MerchantCount
	if nextID == 0 {
		if maxID == 0 && len(genState.MerchantList) == 0 {
			nextID = 1
		} else {
			nextID = maxID + 1
		}
	}
	k.SetNextMerchantID(sdkCtx, nextID)

	return nil
}

// ExportGenesis exporta o estado do módulo para o genesis.
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	genesis := types.DefaultGenesis()

	// Params
	p, err := k.ParamsStore.Get(sdkCtx)
	if err != nil {
		return nil, err
	}
	genesis.Params = p

	// Merchants
	store := prefix.NewStore(k.getStore(sdkCtx), types.MerchantKeyPrefix)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var merchant types.Merchant
		k.cdc.MustUnmarshal(iter.Value(), &merchant)
		genesis.MerchantList = append(genesis.MerchantList, merchant)
	}

	// Sequência
	genesis.MerchantCount = k.GetNextMerchantID(sdkCtx)

	return genesis, nil
}
