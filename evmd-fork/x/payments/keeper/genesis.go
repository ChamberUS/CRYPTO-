package keeper

import (
	"context"

	"cosmossdk.io/collections"

	"github.com/buynnex/iaos-evmd/x/payments/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx context.Context, genState types.GenesisState) error {
	params := genState.Params
	if (params == types.Params{}) {
		params = types.DefaultParams()
	}
	if err := k.Params.Set(ctx, params); err != nil {
		return err
	}

	for _, pr := range genState.PaymentRequests {
		if err := k.PaymentRequests.Set(ctx, pr.Id, pr); err != nil {
			return err
		}
		if err := k.PaymentRequestsByLoja.Set(ctx, collections.Join(pr.LojaId, pr.Id)); err != nil {
			return err
		}
	}

	if err := k.PaymentRequestSeq.Set(ctx, genState.PaymentRequestCount); err != nil {
		return err
	}

	return nil
}

// ExportGenesis returns the module's exported genesis.
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	var err error

	genesis := types.DefaultGenesis()
	genesis.Params, err = k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	err = k.PaymentRequests.Walk(ctx, nil, func(_ uint64, pr types.PaymentRequest) (bool, error) {
		genesis.PaymentRequests = append(genesis.PaymentRequests, pr)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	if genesis.PaymentRequestCount, err = k.PaymentRequestSeq.Peek(ctx); err != nil {
		return nil, err
	}

	return genesis, nil
}
