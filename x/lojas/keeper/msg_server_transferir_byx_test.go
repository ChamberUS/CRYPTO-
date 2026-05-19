package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/buynnex-corp/byx/x/lojas/keeper"
	"github.com/buynnex-corp/byx/x/lojas/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func TestTransferirByx(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)
	sdkCtx := sdk.UnwrapSDKContext(f.ctx)

	owner, err := f.addressCodec.BytesToString([]byte{1, 1, 1, 1, 1})
	require.NoError(t, err)
	operator, err := f.addressCodec.BytesToString([]byte{2, 2, 2, 2, 2})
	require.NoError(t, err)
	other, err := f.addressCodec.BytesToString([]byte{3, 3, 3, 3, 3})
	require.NoError(t, err)
	admin, err := f.addressCodec.BytesToString([]byte{4, 4, 4, 4, 4})
	require.NoError(t, err)

	params := types.DefaultParams()
	params.FaucetAdmin = admin
	require.NoError(t, f.keeper.ParamsStore.Set(sdkCtx, params))

	require.NoError(t, f.keeper.SetMerchant(sdkCtx, types.Merchant{
		Id:       1,
		Creator:  owner,
		Endereco: operator,
		Saldo:    "100",
	}))
	require.NoError(t, f.keeper.SetMerchant(sdkCtx, types.Merchant{
		Id:       2,
		Creator:  other,
		Endereco: other,
		Saldo:    "10",
	}))

	t.Run("dono transfere: passa", func(t *testing.T) {
		_, err := ms.TransferirByx(f.ctx, &types.MsgTransferirByx{
			Creator:       owner,
			DeLojistaId:   "1",
			ParaLojistaId: "2",
			Valor:         "5",
		})
		require.NoError(t, err)
	})

	t.Run("terceiro tenta transferir: falha", func(t *testing.T) {
		_, err := ms.TransferirByx(f.ctx, &types.MsgTransferirByx{
			Creator:       other,
			DeLojistaId:   "1",
			ParaLojistaId: "2",
			Valor:         "1",
		})
		require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)
	})

	t.Run("saldo insuficiente: falha", func(t *testing.T) {
		_, err := ms.TransferirByx(f.ctx, &types.MsgTransferirByx{
			Creator:       owner,
			DeLojistaId:   "1",
			ParaLojistaId: "2",
			Valor:         "999999",
		})
		require.ErrorIs(t, err, sdkerrors.ErrInsufficientFunds)
	})

	t.Run("lojista inexistente: falha", func(t *testing.T) {
		_, err := ms.TransferirByx(f.ctx, &types.MsgTransferirByx{
			Creator:       owner,
			DeLojistaId:   "1",
			ParaLojistaId: "9999",
			Valor:         "1",
		})
		require.ErrorIs(t, err, sdkerrors.ErrKeyNotFound)
	})

	t.Run("fromID == toID: falha", func(t *testing.T) {
		_, err := ms.TransferirByx(f.ctx, &types.MsgTransferirByx{
			Creator:       owner,
			DeLojistaId:   "1",
			ParaLojistaId: "1",
			Valor:         "1",
		})
		require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
	})
}
