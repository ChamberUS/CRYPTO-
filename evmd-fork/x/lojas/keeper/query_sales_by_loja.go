package keeper

import (
	"context"
	"errors"

	"github.com/buynnex/iaos-evmd/x/lojas/types"

	"cosmossdk.io/collections"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GET /byx/lojas/v1/sales_by_loja?loja_id=ID
func (k *Keeper) SalesByLoja(ctx context.Context, req *types.QuerySalesByLojaRequest) (*types.QuerySalesByLojaResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.LojaId == 0 {
		return nil, status.Error(codes.InvalidArgument, "loja_id is required")
	}

	start := req.StartTime
	end := req.EndTime

	pageReq := req.Pagination
	if pageReq == nil {
		pageReq = &sdkquery.PageRequest{Limit: 10, Reverse: true}
	} else if pageReq.Limit == 0 {
		pageReq.Limit = 10
	}

	sales, page, err := sdkquery.CollectionPaginate(
		ctx,
		k.SalesByLojaIndex,
		pageReq,
		func(key collections.Pair[uint64, uint64], _ collections.NoValue) (*types.Sale, error) {
			s, err := k.Sales.Get(ctx, key.K2())
			if err != nil {
				return nil, err
			}
			if start > 0 || end > 0 {
				if s.BlockTime == 0 {
					// vendas antigas sem block_time são ignoradas quando filtros são fornecidos
					return nil, nil
				}
				if start > 0 && s.BlockTime < start {
					return nil, nil
				}
				if end > 0 && s.BlockTime > end {
					return nil, nil
				}
			}
			return &s, nil
		},
		sdkquery.WithCollectionPaginationPairPrefix[uint64, uint64](req.LojaId),
	)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return &types.QuerySalesByLojaResponse{
				Sales:      []*types.Sale{},
				Pagination: page,
			}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	// remove itens filtrados (nil)
	filtered := make([]*types.Sale, 0, len(sales))
	for _, s := range sales {
		if s != nil {
			filtered = append(filtered, s)
		}
	}

	return &types.QuerySalesByLojaResponse{
		Sales:      filtered,
		Pagination: page,
	}, nil
}
