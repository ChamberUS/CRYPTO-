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
		Creator:         creator,
		Nome:            "Loja Segura",
		Endereco:        "Rua A, 123",
		OperatorAddress: creator,
		KycRef:          "kyc-local-1",
		DocumentHash:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		KycStatus:       "approved",
	})
	require.NoError(t, err)

	created, err := f.keeper.GetMerchant(f.ctx, createResp.Id)
	require.NoError(t, err)
	require.Equal(t, "0", created.Saldo, "create must always initialize saldo to zero")
	require.Equal(t, creator, created.OperatorAddress)
	require.Equal(t, "kyc-local-1", created.KycRef)
	require.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", created.DocumentHash)
	require.Equal(t, "approved", created.KycStatus)

	created.Saldo = "77"
	require.NoError(t, f.keeper.SetMerchant(sdk.UnwrapSDKContext(f.ctx), created))

	_, err = srv.UpdateMerchant(f.ctx, &types.MsgUpdateMerchant{
		Creator:         creator,
		Id:              createResp.Id,
		Nome:            "Loja Segura 2",
		Endereco:        "Rua B, 999",
		OperatorAddress: creator,
		KycRef:          "kyc-local-2",
		DocumentHash:    "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		KycStatus:       "rejected",
	})
	require.NoError(t, err)

	updated, err := f.keeper.GetMerchant(f.ctx, createResp.Id)
	require.NoError(t, err)
	require.Equal(t, "77", updated.Saldo, "update must preserve existing saldo")
	require.Equal(t, "kyc-local-2", updated.KycRef)
	require.Equal(t, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", updated.DocumentHash)
	require.Equal(t, "rejected", updated.KycStatus)
}

func TestMerchantValidationOperatorKYCAndHash(t *testing.T) {
	f := initFixture(t)
	srv := keeper.NewMsgServerImpl(f.keeper)

	creator, err := f.addressCodec.BytesToString([]byte("owner_validation___________"))
	require.NoError(t, err)

	t.Run("invalid operator address fails", func(t *testing.T) {
		_, err := srv.CreateMerchant(f.ctx, &types.MsgCreateMerchant{
			Creator:         creator,
			Nome:            "Loja",
			Endereco:        "Rua X",
			OperatorAddress: "invalid",
		})
		require.ErrorIs(t, err, sdkerrors.ErrInvalidAddress)
	})

	t.Run("invalid kyc status fails", func(t *testing.T) {
		_, err := srv.CreateMerchant(f.ctx, &types.MsgCreateMerchant{
			Creator:   creator,
			Nome:      "Loja",
			Endereco:  "Rua X",
			KycStatus: "unknown",
		})
		require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
	})

	t.Run("invalid document hash fails", func(t *testing.T) {
		_, err := srv.CreateMerchant(f.ctx, &types.MsgCreateMerchant{
			Creator:      creator,
			Nome:         "Loja",
			Endereco:     "Rua X",
			DocumentHash: "abc",
		})
		require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
	})

	t.Run("empty operator allowed", func(t *testing.T) {
		resp, err := srv.CreateMerchant(f.ctx, &types.MsgCreateMerchant{
			Creator:      creator,
			Nome:         "Loja Ok",
			Endereco:     "Rua Y",
			KycStatus:    "pending",
			DocumentHash: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		})
		require.NoError(t, err)

		m, err := f.keeper.GetMerchant(f.ctx, resp.Id)
		require.NoError(t, err)
		require.Equal(t, "", m.OperatorAddress)
	})
}
