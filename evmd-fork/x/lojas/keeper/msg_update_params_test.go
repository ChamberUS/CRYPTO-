package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/buynnex/iaos-evmd/x/lojas/keeper"
	"github.com/buynnex/iaos-evmd/x/lojas/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgUpdateParams(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)

	params := types.DefaultParams()
	require.NoError(t, f.keeper.ParamsStore.Set(f.ctx, params))

	authorityStr, err := f.addressCodec.BytesToString(f.keeper.GetAuthority())
	require.NoError(t, err)

	testCases := []struct {
		name      string
		input     *types.MsgUpdateParams
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid authority",
			input: &types.MsgUpdateParams{
				Authority: "invalid",
				Params:    params,
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "invalid params rejected",
			input: &types.MsgUpdateParams{
				Authority: authorityStr,
				Params: types.Params{
					FaucetEnabled:           true,
					MaxValorVendaEmCentavos: 0, // invalid
					MaxCashbackUbyxPorVenda: 0,
				},
			},
			expErr:    true,
			expErrMsg: "invalid request",
		},
		{
			name: "all good",
			input: &types.MsgUpdateParams{
				Authority: authorityStr,
				Params:    params,
			},
			expErr: false,
		},
		{
			name: "emits diff when params change",
			input: &types.MsgUpdateParams{
				Authority: authorityStr,
				Params: func() types.Params {
					p := params
					p.FaucetEnabled = !p.FaucetEnabled
					p.MaxSalesPerBlockPerLoja = 30
					return p
				}(),
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sdkCtx := sdk.UnwrapSDKContext(f.ctx)
			ctx := sdkCtx.WithEventManager(sdk.NewEventManager())
			_, err := ms.UpdateParams(ctx, tc.input)

			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
				if tc.name == "emits diff when params change" {
					events := ctx.EventManager().Events()
					require.True(t, len(events) > 0)
					found := false
					for _, ev := range events {
						if ev.Type == "lojas_params_updated" {
							for _, attr := range ev.Attributes {
								if string(attr.Key) == "param_name" && string(attr.Value) == "faucet_enabled" {
									found = true
								}
							}
						}
					}
					require.True(t, found, "expected diff attributes for params change")
				}
			}
		})
	}
}
