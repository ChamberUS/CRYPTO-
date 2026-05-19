package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/buynnex-corp/byx/x/lojas/keeper"
	"github.com/buynnex-corp/byx/x/lojas/types"
)

func TestFaucetRespectsAdmin(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)

	adminBytes := make([]byte, 20)
	adminBytes[0] = 1
	adminStr, err := f.addressCodec.BytesToString(adminBytes)
	require.NoError(t, err)

	params := types.DefaultParams()
	params.FaucetAdmin = adminStr
	require.NoError(t, f.keeper.ParamsStore.Set(f.ctx, params))

	otherBytes := make([]byte, 20)
	otherBytes[0] = 2
	otherStr, err := f.addressCodec.BytesToString(otherBytes)
	require.NoError(t, err)

	_, err = ms.Faucet(f.ctx, &types.MsgFaucet{
		Creator:   otherStr,
		LojistaId: "1",
		Amount:    "10",
	})
	require.Error(t, err, "should block unauthorized faucet caller")
}

func TestFaucetRequiresAdminWhenEnabled(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)

	params := types.DefaultParams()
	params.FaucetEnabled = true
	params.FaucetAdmin = ""
	require.NoError(t, f.keeper.ParamsStore.Set(f.ctx, params))

	creatorBytes := make([]byte, 20)
	creatorBytes[0] = 9
	creator, err := f.addressCodec.BytesToString(creatorBytes)
	require.NoError(t, err)

	_, err = ms.Faucet(f.ctx, &types.MsgFaucet{
		Creator:   creator,
		LojistaId: "1",
		Amount:    "10",
	})
	require.Error(t, err)
}
