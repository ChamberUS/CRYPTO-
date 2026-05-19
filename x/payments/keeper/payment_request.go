package keeper

import (
	"encoding/binary"

	"github.com/buynnex-corp/byx/x/payments/types"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) getStore(ctx sdk.Context) storetypes.KVStore {
	return runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
}

func (k Keeper) GetNextPaymentRequestID(ctx sdk.Context) uint64 {
	store := k.getStore(ctx)
	bz := store.Get(types.NextPaymentRequestIDKey)
	if bz == nil {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

func (k Keeper) SetNextPaymentRequestID(ctx sdk.Context, id uint64) {
	store := k.getStore(ctx)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	store.Set(types.NextPaymentRequestIDKey, bz)
}

func (k Keeper) SetPaymentRequest(ctx sdk.Context, pr types.PaymentRequest) {
	store := k.getStore(ctx)
	key := append(types.PaymentRequestKeyPrefix, sdk.Uint64ToBigEndian(pr.Id)...)
	bz := k.cdc.MustMarshal(&pr)
	store.Set(key, bz)
}

func (k Keeper) GetPaymentRequest(ctx sdk.Context, id uint64) (types.PaymentRequest, bool) {
	store := k.getStore(ctx)
	key := append(types.PaymentRequestKeyPrefix, sdk.Uint64ToBigEndian(id)...)
	bz := store.Get(key)
	if bz == nil {
		return types.PaymentRequest{}, false
	}
	var pr types.PaymentRequest
	k.cdc.MustUnmarshal(bz, &pr)
	return pr, true
}

func (k Keeper) AddPaymentRequestToLojaIndex(ctx sdk.Context, lojaID uint64, requestID uint64) {
	store := k.getStore(ctx)
	prefixStore := prefix.NewStore(store, types.PaymentRequestByLojaPrefix)
	lojaStore := prefix.NewStore(prefixStore, sdk.Uint64ToBigEndian(lojaID))
	key := sdk.Uint64ToBigEndian(requestID)
	lojaStore.Set(key, []byte{})
}
