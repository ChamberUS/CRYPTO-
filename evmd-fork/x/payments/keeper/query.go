package keeper

import (
	"context"
	"errors"

	"github.com/buynnex/iaos-evmd/x/payments/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = queryServer{}

// NewQueryServerImpl returns an implementation of the QueryServer interface
// for the provided Keeper.
func NewQueryServerImpl(k Keeper) types.QueryServer {
	return queryServer{k}
}

type queryServer struct {
	k Keeper
}

func (q queryServer) PaymentRequest(ctx context.Context, req *types.QueryGetPaymentRequestRequest) (*types.QueryGetPaymentRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pr, found := q.k.GetPaymentRequest(sdkCtx, req.Id)
	if !found {
		return nil, status.Error(codes.NotFound, "payment request not found")
	}

	if err := q.k.ensureCurrentStatus(sdkCtx, &pr); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryGetPaymentRequestResponse{PaymentRequest: &pr}, nil
}

func (q queryServer) PaymentRequestsByLoja(ctx context.Context, req *types.QueryPaymentRequestsByLojaRequest) (*types.QueryPaymentRequestsByLojaResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.LojaId == 0 {
		return nil, status.Error(codes.InvalidArgument, "loja_id is required")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pageReq := req.Pagination
	if pageReq == nil {
		pageReq = &sdkquery.PageRequest{Limit: 10, Reverse: true}
	} else if pageReq.Limit == 0 {
		pageReq.Limit = 10
	}

	baseStore := runtime.KVStoreAdapter(q.k.storeService.OpenKVStore(sdkCtx))
	pfxStore := prefix.NewStore(baseStore, types.PaymentRequestByLojaPrefix)
	lojaStore := prefix.NewStore(pfxStore, sdk.Uint64ToBigEndian(req.LojaId))

	var items []*types.PaymentRequest
	page, err := sdkquery.Paginate(lojaStore, pageReq, func(key []byte, _ []byte) error {
		id := sdk.BigEndianToUint64(key)
		pr, found := q.k.GetPaymentRequest(sdkCtx, id)
		if !found {
			return collections.ErrNotFound
		}
		if err := q.k.ensureCurrentStatus(sdkCtx, &pr); err != nil {
			return err
		}
		items = append(items, &pr)
		return nil
	})
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPaymentRequestsByLojaResponse{
		PaymentRequests: items,
		Pagination:      page,
	}, nil
}
