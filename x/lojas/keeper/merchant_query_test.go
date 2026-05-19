package keeper_test

import (
	"testing"

	"github.com/buynnex-corp/byx/x/lojas/keeper"
	"github.com/buynnex-corp/byx/x/lojas/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestCreateMerchantAllocatesID(t *testing.T) {
	f := initFixture(t)
	srv := keeper.NewMsgServerImpl(f.keeper)

	creator, err := f.addressCodec.BytesToString([]byte("merchant_creator___________"))
	require.NoError(t, err)

	resp, err := srv.CreateMerchant(f.ctx, &types.MsgCreateMerchant{
		Creator: creator,
		Nome:    "Loja Teste",
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), resp.Id)

	merchant, err := f.keeper.GetMerchant(f.ctx, resp.Id)
	require.NoError(t, err)
	require.Equal(t, resp.Id, merchant.Id)
	require.Equal(t, creator, merchant.Creator)
}

func TestQueryMerchantByID(t *testing.T) {
	f := initFixture(t)
	sdkCtx := sdk.UnwrapSDKContext(f.ctx)

	merchant := types.Merchant{Id: 7, Nome: "Query", Creator: "creator", Cpfcnpj: "123", Telefone: "999"}
	require.NoError(t, f.keeper.SetMerchant(sdkCtx, merchant))

	resp, err := f.keeper.Merchant(sdk.WrapSDKContext(sdkCtx), &types.QueryGetMerchantRequest{Id: merchant.Id})
	require.NoError(t, err)
	require.Equal(t, merchant.Id, resp.Merchant.Id)
	require.Equal(t, "", resp.Merchant.Cpfcnpj)
	require.Equal(t, "", resp.Merchant.Telefone)
}

func TestMerchantAllReturnsID(t *testing.T) {
	f := initFixture(t)
	sdkCtx := sdk.UnwrapSDKContext(f.ctx)

	merchants := []types.Merchant{
		{Id: 3, Nome: "Loja1", Cpfcnpj: "123", Telefone: "111"},
		{Id: 4, Nome: "Loja2", Cpfcnpj: "456", Telefone: "222"},
	}
	for _, m := range merchants {
		require.NoError(t, f.keeper.SetMerchant(sdkCtx, m))
	}

	resp, err := f.keeper.MerchantAll(sdk.WrapSDKContext(sdkCtx), &types.QueryAllMerchantRequest{})
	require.NoError(t, err)
	require.Len(t, resp.Merchant, len(merchants))
	for _, m := range resp.Merchant {
		require.NotZero(t, m.Id)
		require.Equal(t, "", m.Cpfcnpj)
		require.Equal(t, "", m.Telefone)
	}
}
