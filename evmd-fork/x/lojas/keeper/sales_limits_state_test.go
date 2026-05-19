package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/collections"
	"github.com/stretchr/testify/require"

	"github.com/buynnex/iaos-evmd/x/lojas/keeper"
	"github.com/buynnex/iaos-evmd/x/lojas/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestSalesCountStateResetsPerHeight(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)

	creator, err := f.addressCodec.BytesToString([]byte{9, 9, 9})
	require.NoError(t, err)

	merchant := types.Merchant{Id: 10, Nome: "H", Endereco: creator, Creator: creator}
	sdkCtx := sdk.UnwrapSDKContext(f.ctx)
	require.NoError(t, f.keeper.SetMerchant(sdkCtx, merchant))
	f.keeper.SetNextMerchantID(sdkCtx, merchant.Id+1)
	f.keeper.SetMerchantByCreator(sdkCtx, creator, merchant.Id)

	params := types.DefaultParams()
	params.MaxSalesPerBlockPerLoja = 1
	require.NoError(t, f.keeper.ParamsStore.Set(sdkCtx, params))

	// First height -> ok
	_, err = ms.RegistrarVenda(sdk.WrapSDKContext(sdkCtx.WithBlockHeight(100)), &types.MsgRegistrarVenda{
		Creator:         creator,
		LojaId:          "10",
		ValorEmCentavos: 1000,
	})
	require.NoError(t, err)

	// Same height -> exceed
	_, err = ms.RegistrarVenda(sdk.WrapSDKContext(sdkCtx.WithBlockHeight(100)), &types.MsgRegistrarVenda{
		Creator:         creator,
		LojaId:          "10",
		ValorEmCentavos: 1000,
	})
	require.Error(t, err)

	// New height resets
	_, err = ms.RegistrarVenda(sdk.WrapSDKContext(sdkCtx.WithBlockHeight(101)), &types.MsgRegistrarVenda{
		Creator:         creator,
		LojaId:          "10",
		ValorEmCentavos: 1000,
	})
	require.NoError(t, err)
}

func TestDailyRetentionDeletesOldKey(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)

	creator, err := f.addressCodec.BytesToString([]byte{7, 7, 7})
	require.NoError(t, err)
	merchant := types.Merchant{Id: 11, Nome: "Ret", Endereco: creator, Creator: creator}
	sdkCtx := sdk.UnwrapSDKContext(f.ctx)
	require.NoError(t, f.keeper.SetMerchant(sdkCtx, merchant))
	f.keeper.SetNextMerchantID(sdkCtx, merchant.Id+1)
	f.keeper.SetMerchantByCreator(sdkCtx, creator, merchant.Id)

	params := types.DefaultParams()
	params.CashbackDailyRetentionDays = 2
	require.NoError(t, f.keeper.ParamsStore.Set(sdkCtx, params))

	_ = sdkCtx.WithBlockTime(time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC))
	oldKey := uint64(20240101)
	require.NoError(t, f.keeper.DailyCashbackByLoja.Set(f.ctx, collections.Join(uint64(11), oldKey), 100))

	// Advance beyond retention; registering new sale should prune old key
	day3 := sdkCtx.WithBlockTime(time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC))
	_, err = ms.RegistrarVenda(sdk.WrapSDKContext(day3), &types.MsgRegistrarVenda{
		Creator:         creator,
		LojaId:          "11",
		ValorEmCentavos: 1000,
		Cliente:         "",
	})
	require.NoError(t, err)

	_, err = f.keeper.DailyCashbackByLoja.Get(f.ctx, collections.Join(uint64(11), oldKey))
	require.Error(t, err, "expected old accumulator removed")
}
