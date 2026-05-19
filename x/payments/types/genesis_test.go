package types_test

import (
	"testing"

	"github.com/buynnex-corp/byx/x/payments/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
	tests := []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "duplicate payment request ids",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				PaymentRequests: []types.PaymentRequest{
					{Id: 1, LojaId: 10},
					{Id: 1, LojaId: 11},
				},
			},
			valid: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
