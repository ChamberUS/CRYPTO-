package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"github.com/buynnex/iaos-evmd/x/lojas/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GET /byx/lojas/v1/sales_summary?loja_id=ID
func (k *Keeper) SalesSummary(ctx context.Context, req *types.QuerySalesSummaryRequest) (*types.QuerySalesSummaryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	iter, err := k.SalesByLojaIndex.Iterate(ctx, collections.NewPrefixedPairRange[uint64, uint64](req.LojaId))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer iter.Close()

	var (
		totalVendas           uint64
		totalValorEmCentavos  uint64
		totalCashbackMicroByx uint64
	)

	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		sale, err := k.Sales.Get(ctx, key.K2())
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		totalVendas++
		totalValorEmCentavos += sale.ValorEmCentavos
		totalCashbackMicroByx += sale.CashbackMicroByx
	}

	var ticketMedio uint64
	if totalVendas > 0 {
		ticketMedio = totalValorEmCentavos / totalVendas
	}

	return &types.QuerySalesSummaryResponse{
		Summary: &types.SalesSummary{
			LojaId:                req.LojaId,
			TotalVendas:           totalVendas,
			TotalValorEmCentavos:  totalValorEmCentavos,
			TotalCashbackMicroByx: totalCashbackMicroByx,
			TicketMedioEmCentavos: ticketMedio,
		},
	}, nil
}
