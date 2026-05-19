package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/buynnex-corp/byx/x/lojas/types"
)

func (k *Keeper) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	p, err := k.ParamsStore.Get(sdk.UnwrapSDKContext(ctx))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// proto espera *Params
	return &types.QueryParamsResponse{Params: &p}, nil
}
