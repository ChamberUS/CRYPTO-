package keeper_test

import (
	"testing"

	"github.com/buynnex/iaos-evmd/x/lojas/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params:        types.DefaultParams(),
		MerchantList:  []types.Merchant{{Id: 1}, {Id: 2}},
		MerchantCount: 3,
	}
	f := initFixture(t)
	err := f.keeper.InitGenesis(f.ctx, genesisState)
	require.NoError(t, err)
	got, err := f.keeper.ExportGenesis(f.ctx)
	require.NoError(t, err)
	require.NotNil(t, got)

	require.EqualExportedValues(t, genesisState.Params, got.Params)
	require.EqualExportedValues(t, genesisState.MerchantList, got.MerchantList)
	require.Equal(t, genesisState.MerchantCount, got.MerchantCount)

}
