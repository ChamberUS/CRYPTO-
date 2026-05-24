package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/buynnex-corp/byx/x/lojas/keeper"
	"github.com/buynnex-corp/byx/x/lojas/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestRegistrarVenda_DailyLimitExceeded(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)

	creator, err := f.addressCodec.BytesToString([]byte{1, 2, 3, 4, 5})
	require.NoError(t, err)

	merchant := types.Merchant{
		Id:       1,
		Nome:     "Loja",
		Endereco: creator,
		Saldo:    "0",
		Creator:  creator,
	}
	sdkCtx := sdk.UnwrapSDKContext(f.ctx)
	require.NoError(t, f.keeper.SetMerchant(sdkCtx, merchant))
	f.keeper.SetNextMerchantID(sdkCtx, merchant.Id+1)
	f.keeper.SetMerchantByCreator(sdkCtx, creator, merchant.Id)

	params := types.DefaultParams()
	params.LimitsEnabled = true
	params.MaxCashbackDailyPerLojaUbyx = 5
	params.MaxCashbackUbyxPorVenda = 500
	params.MaxSalesPerBlockPerLoja = 5
	params.CashbackRateUbyxPerReal = 10
	require.NoError(t, f.keeper.ParamsStore.Set(sdkCtx, params))

	// cashback calculado será > limite diário e deve falhar antes do mint
	_, err = ms.RegistrarVenda(sdk.WrapSDKContext(sdkCtx), &types.MsgRegistrarVenda{
		Creator:         creator,
		LojaId:          "1",
		ValorEmCentavos: 2000, // 20 reais -> 200 micro com rate 10
		Cliente:         creator,
	})
	require.Error(t, err)
}

func TestRegistrarVenda_BlockLimitExceeded(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)

	creator, err := f.addressCodec.BytesToString([]byte{1, 2, 3, 4, 5})
	require.NoError(t, err)

	merchant := types.Merchant{
		Id:       2,
		Nome:     "LojaB",
		Endereco: creator,
		Saldo:    "0",
		Creator:  creator,
	}
	sdkCtx := sdk.UnwrapSDKContext(f.ctx)
	require.NoError(t, f.keeper.SetMerchant(sdkCtx, merchant))
	f.keeper.SetNextMerchantID(sdkCtx, merchant.Id+1)
	f.keeper.SetMerchantByCreator(sdkCtx, creator, merchant.Id)

	params := types.DefaultParams()
	params.LimitsEnabled = true
	params.MaxSalesPerBlockPerLoja = 1
	require.NoError(t, f.keeper.ParamsStore.Set(sdkCtx, params))

	// primeiro passa (cliente vazio evita mint)
	_, err = ms.RegistrarVenda(sdk.WrapSDKContext(sdkCtx), &types.MsgRegistrarVenda{
		Creator:         creator,
		LojaId:          "2",
		ValorEmCentavos: 1000,
		Cliente:         "",
	})
	require.NoError(t, err)

	// segunda no mesmo bloco deve falhar
	_, err = ms.RegistrarVenda(sdk.WrapSDKContext(sdkCtx), &types.MsgRegistrarVenda{
		Creator:         creator,
		LojaId:          "2",
		ValorEmCentavos: 1000,
		Cliente:         "",
	})
	require.Error(t, err)
}

func TestRegistrarVenda_MaxValorExceeded(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)

	creator, err := f.addressCodec.BytesToString([]byte{9, 8, 7, 6, 5})
	require.NoError(t, err)

	merchant := types.Merchant{
		Id:       3,
		Nome:     "Loja Limite",
		Endereco: creator,
		Saldo:    "0",
		Creator:  creator,
	}
	sdkCtx := sdk.UnwrapSDKContext(f.ctx)
	require.NoError(t, f.keeper.SetMerchant(sdkCtx, merchant))
	f.keeper.SetNextMerchantID(sdkCtx, merchant.Id+1)
	f.keeper.SetMerchantByCreator(sdkCtx, creator, merchant.Id)

	params := types.DefaultParams()
	params.MaxValorVendaEmCentavos = 1_500
	require.NoError(t, f.keeper.ParamsStore.Set(sdkCtx, params))

	_, err = ms.RegistrarVenda(sdk.WrapSDKContext(sdkCtx), &types.MsgRegistrarVenda{
		Creator:         creator,
		LojaId:          "3",
		ValorEmCentavos: 2_000,
		Cliente:         "",
	})
	require.Error(t, err)
}
