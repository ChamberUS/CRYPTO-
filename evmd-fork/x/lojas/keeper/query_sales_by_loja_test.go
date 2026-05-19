package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/collections"
	"github.com/stretchr/testify/require"

	"github.com/buynnex/iaos-evmd/x/lojas/types"
)

func TestSalesByLojaFiltersBlockTimeZeroWhenRange(t *testing.T) {
	f := initFixture(t)
	creator, err := f.addressCodec.BytesToString([]byte{1, 2, 3})
	require.NoError(t, err)

	// two sales: one with block_time zero, one with block_time set
	s1 := types.Sale{Id: 1, LojaId: 1, Creator: creator, ValorEmCentavos: 100}
	s2 := types.Sale{Id: 2, LojaId: 1, Creator: creator, ValorEmCentavos: 200, BlockTime: uint64(time.Now().Add(-time.Hour).Unix())}

	require.NoError(t, f.keeper.Sales.Set(f.ctx, s1.Id, s1))
	require.NoError(t, f.keeper.Sales.Set(f.ctx, s2.Id, s2))
	require.NoError(t, f.keeper.SalesByLojaIndex.Set(f.ctx, collections.Join(s1.LojaId, s1.Id)))
	require.NoError(t, f.keeper.SalesByLojaIndex.Set(f.ctx, collections.Join(s2.LojaId, s2.Id)))

	// with filter -> should ignore block_time zero
	res, err := f.keeper.SalesByLoja(f.ctx, &types.QuerySalesByLojaRequest{
		LojaId:    1,
		StartTime: uint64(time.Now().Add(-2 * time.Hour).Unix()),
		EndTime:   uint64(time.Now().Unix()),
	})
	require.NoError(t, err)
	require.Len(t, res.Sales, 1)
	require.Equal(t, s2.Id, res.Sales[0].Id)

	// without filter -> include all
	resAll, err := f.keeper.SalesByLoja(f.ctx, &types.QuerySalesByLojaRequest{LojaId: 1})
	require.NoError(t, err)
	require.Len(t, resAll.Sales, 2)
}
