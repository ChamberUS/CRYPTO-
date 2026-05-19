package keeper

import (
	"context"
	"encoding/binary"

	"github.com/buynnex-corp/byx/x/lojas/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) getStore(ctx sdk.Context) storetypes.KVStore {
	return runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
}

func (k Keeper) merchantKey(id uint64) []byte {
	return append(types.MerchantKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func (k Keeper) GetNextMerchantID(ctx sdk.Context) uint64 {
	store := k.getStore(ctx)
	bz := store.Get(types.NextMerchantIDKey)
	if bz == nil {
		// Reserve 0 as invalid to avoid clashes with payments that require loja_id > 0.
		return 1
	}
	return binary.BigEndian.Uint64(bz)
}

func (k Keeper) SetNextMerchantID(ctx sdk.Context, id uint64) {
	store := k.getStore(ctx)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	store.Set(types.NextMerchantIDKey, bz)
}

func (k Keeper) AllocateMerchantID(ctx sdk.Context) uint64 {
	nextID := k.GetNextMerchantID(ctx)
	if nextID == 0 {
		nextID = 1
	}
	k.SetNextMerchantID(ctx, nextID+1)
	return nextID
}

func (k Keeper) SetMerchant(ctx sdk.Context, merchant types.Merchant) error {
	store := k.getStore(ctx)
	store.Set(k.merchantKey(merchant.Id), k.cdc.MustMarshal(&merchant))
	next := k.GetNextMerchantID(ctx)
	if merchant.Id >= next {
		k.SetNextMerchantID(ctx, merchant.Id+1)
	}
	return nil
}

func (k Keeper) getMerchant(ctx sdk.Context, id uint64) (types.Merchant, bool) {
	store := k.getStore(ctx)
	bz := store.Get(k.merchantKey(id))
	if bz == nil {
		return types.Merchant{}, false
	}
	var merchant types.Merchant
	k.cdc.MustUnmarshal(bz, &merchant)
	return merchant, true
}

func (k Keeper) GetMerchant(ctx context.Context, id uint64) (types.Merchant, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	merchant, found := k.getMerchant(sdkCtx, id)
	if !found {
		return types.Merchant{}, collections.ErrNotFound
	}
	return merchant, nil
}

func (k Keeper) SetMerchantByCreator(ctx sdk.Context, creator string, id uint64) {
	store := prefix.NewStore(k.getStore(ctx), types.MerchantByCreatorPrefix)
	store.Set([]byte(creator), sdk.Uint64ToBigEndian(id))
}

func (k Keeper) GetMerchantIDByCreator(ctx sdk.Context, creator string) (uint64, bool) {
	store := prefix.NewStore(k.getStore(ctx), types.MerchantByCreatorPrefix)
	bz := store.Get([]byte(creator))
	if bz == nil {
		return 0, false
	}
	return binary.BigEndian.Uint64(bz), true
}
