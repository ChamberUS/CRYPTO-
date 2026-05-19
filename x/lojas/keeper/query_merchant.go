package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/buynnex-corp/byx/x/lojas/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func redactMerchantPublic(m types.Merchant) types.Merchant {
	return m
}

// GET /byx/lojas/v1/merchant
func (k *Keeper) MerchantAll(ctx context.Context, req *types.QueryAllMerchantRequest) (*types.QueryAllMerchantResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(k.getStore(sdkCtx), types.MerchantKeyPrefix)
	var merchants []*types.Merchant

	pageRes, err := sdkquery.Paginate(store, req.Pagination, func(_ []byte, value []byte) error {
		var merchant types.Merchant
		k.cdc.MustUnmarshal(value, &merchant)
		merchant = redactMerchantPublic(merchant)
		merchants = append(merchants, &merchant)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryAllMerchantResponse{Merchant: merchants, Pagination: pageRes}, nil
}

// GET /byx/lojas/v1/merchant/{id}
func (k *Keeper) Merchant(ctx context.Context, req *types.QueryGetMerchantRequest) (*types.QueryGetMerchantResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	m, found := k.getMerchant(sdkCtx, req.Id)
	if !found {
		return nil, status.Error(codes.NotFound, "merchant not found")
	}
	m = redactMerchantPublic(m)
	return &types.QueryGetMerchantResponse{Merchant: &m}, nil
}
