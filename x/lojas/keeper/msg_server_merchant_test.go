package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/buynnex-corp/byx/x/lojas/keeper"
	"github.com/buynnex-corp/byx/x/lojas/types"
)

func TestMerchantMsgServerCreate(t *testing.T) {
	f := initFixture(t)
	srv := keeper.NewMsgServerImpl(f.keeper)

	creator, err := f.addressCodec.BytesToString([]byte("signerAddr__________________"))
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		resp, err := srv.CreateMerchant(f.ctx, &types.MsgCreateMerchant{Creator: creator})
		require.NoError(t, err)
		require.Equal(t, i+1, int(resp.Id))
	}
}

func TestMerchantMsgServerUpdate(t *testing.T) {
	f := initFixture(t)
	srv := keeper.NewMsgServerImpl(f.keeper)

	creator, err := f.addressCodec.BytesToString([]byte("signerAddr__________________"))
	require.NoError(t, err)

	unauthorizedAddr, err := f.addressCodec.BytesToString([]byte("unauthorizedAddr___________"))
	require.NoError(t, err)

	_, err = srv.CreateMerchant(f.ctx, &types.MsgCreateMerchant{Creator: creator})
	require.NoError(t, err)

	tests := []struct {
		desc    string
		request *types.MsgUpdateMerchant
		err     error
	}{
		{
			desc:    "invalid address",
			request: &types.MsgUpdateMerchant{Creator: "invalid"},
			err:     sdkerrors.ErrInvalidAddress,
		},
		{
			desc:    "unauthorized",
			request: &types.MsgUpdateMerchant{Creator: unauthorizedAddr, Id: 1},
			err:     sdkerrors.ErrUnauthorized,
		},
		{
			desc:    "key not found",
			request: &types.MsgUpdateMerchant{Creator: creator, Id: 10},
			err:     sdkerrors.ErrKeyNotFound,
		},
		{
			desc:    "completed",
			request: &types.MsgUpdateMerchant{Creator: creator, Id: 1},
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			_, err = srv.UpdateMerchant(f.ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMerchantMsgServerDelete(t *testing.T) {
	f := initFixture(t)
	srv := keeper.NewMsgServerImpl(f.keeper)

	creator, err := f.addressCodec.BytesToString([]byte("signerAddr__________________"))
	require.NoError(t, err)

	unauthorizedAddr, err := f.addressCodec.BytesToString([]byte("unauthorizedAddr___________"))
	require.NoError(t, err)

	_, err = srv.CreateMerchant(f.ctx, &types.MsgCreateMerchant{Creator: creator})
	require.NoError(t, err)

	tests := []struct {
		desc    string
		request *types.MsgDeleteMerchant
		err     error
	}{
		{
			desc:    "invalid address",
			request: &types.MsgDeleteMerchant{Creator: "invalid"},
			err:     sdkerrors.ErrInvalidAddress,
		},
		{
			desc:    "unauthorized",
			request: &types.MsgDeleteMerchant{Creator: unauthorizedAddr, Id: 1},
			err:     sdkerrors.ErrUnauthorized,
		},
		{
			desc:    "key not found",
			request: &types.MsgDeleteMerchant{Creator: creator, Id: 10},
			err:     sdkerrors.ErrKeyNotFound,
		},
		{
			desc:    "completed",
			request: &types.MsgDeleteMerchant{Creator: creator, Id: 1},
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			_, err = srv.DeleteMerchant(f.ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMerchantSaldoAndPIIHardening(t *testing.T) {
	f := initFixture(t)
	srv := keeper.NewMsgServerImpl(f.keeper)

	creator, err := f.addressCodec.BytesToString([]byte("ownerAddr___________________"))
	require.NoError(t, err)

	createResp, err := srv.CreateMerchant(f.ctx, &types.MsgCreateMerchant{
		Creator:  creator,
		Nome:     "Loja Segura",
		Endereco: creator,
		Cpfcnpj:  "12345678901",
		Telefone: "11999999999",
		Saldo:    "999999",
	})
	require.NoError(t, err)

	created, err := f.keeper.GetMerchant(f.ctx, createResp.Id)
	require.NoError(t, err)
	require.Equal(t, "0", created.Saldo, "create must ignore user-provided saldo")
	require.Equal(t, "", created.Cpfcnpj, "cpfcnpj must not be persisted")
	require.Equal(t, "", created.Telefone, "telefone must not be persisted")

	created.Saldo = "77"
	require.NoError(t, f.keeper.SetMerchant(sdk.UnwrapSDKContext(f.ctx), created))

	_, err = srv.UpdateMerchant(f.ctx, &types.MsgUpdateMerchant{
		Creator:  creator,
		Id:       createResp.Id,
		Nome:     "Loja Segura 2",
		Endereco: creator,
		Cpfcnpj:  "00000000000",
		Telefone: "11888888888",
		Saldo:    "123456",
	})
	require.NoError(t, err)

	updated, err := f.keeper.GetMerchant(f.ctx, createResp.Id)
	require.NoError(t, err)
	require.Equal(t, "77", updated.Saldo, "update must preserve existing saldo")
	require.Equal(t, "", updated.Cpfcnpj, "cpfcnpj must remain redacted")
	require.Equal(t, "", updated.Telefone, "telefone must remain redacted")
}
