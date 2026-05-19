package keeper

import (
	"context"

	"cosmossdk.io/collections"

	"github.com/buynnex-corp/byx/x/certificados/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx context.Context, genState types.GenesisState) error {
	params := genState.Params
	if (params == types.Params{}) {
		params = types.DefaultParams()
	}
	if err := k.ParamsStore.Set(ctx, params); err != nil {
		return err
	}

	for _, cert := range genState.Certificates {
		if err := k.storeCertificate(ctx, cert); err != nil {
			return err
		}
	}

	if genState.CertificateCount > 0 {
		if err := k.CertificateCount.Set(ctx, genState.CertificateCount); err != nil {
			return err
		}
	}

	for _, record := range genState.ServiceRecords {
		if err := k.ServiceRecords.Set(ctx, collections.Join(record.CertificateId, record.Index), record); err != nil {
			return err
		}
	}

	for _, sc := range genState.ServiceCounts {
		if err := k.ServiceRecordsCount.Set(ctx, sc.CertificateId, sc.Count); err != nil {
			return err
		}
	}

	return nil
}

// ExportGenesis returns the module's exported genesis.
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	var err error

	genesis := types.DefaultGenesis()
	genesis.Params = k.ensureParams(ctx)

	err = k.Certificates.Walk(ctx, nil, func(_ uint64, cert types.Certificate) (bool, error) {
		genesis.Certificates = append(genesis.Certificates, cert)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	err = k.ServiceRecords.Walk(ctx, nil, func(_ collections.Pair[uint64, uint64], record types.ServiceRecord) (bool, error) {
		genesis.ServiceRecords = append(genesis.ServiceRecords, record)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	err = k.ServiceRecordsCount.Walk(ctx, nil, func(certID uint64, count uint64) (bool, error) {
		genesis.ServiceCounts = append(genesis.ServiceCounts, types.ServiceCount{
			CertificateId: certID,
			Count:         count,
		})
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	if genesis.CertificateCount, err = k.CertificateCount.Get(ctx); err != nil && err != collections.ErrNotFound {
		return nil, err
	} else if err == collections.ErrNotFound {
		genesis.CertificateCount = 0
	}

	return genesis, nil
}
