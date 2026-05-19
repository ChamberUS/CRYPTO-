package types_test

import (
	"testing"

	"github.com/buynnex-corp/byx/x/lojas/types"

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
			desc:     "valid genesis state",
			genState: &types.GenesisState{MerchantList: []types.Merchant{{Id: 1}, {Id: 2}}, MerchantCount: 3}, valid: true,
		}, {
			desc: "duplicated merchant",
			genState: &types.GenesisState{
				MerchantList: []types.Merchant{
					{
						Id: 1,
					},
					{
						Id: 1,
					},
				},
			},
			valid: false,
		}, {
			desc: "invalid merchant count",
			genState: &types.GenesisState{
				MerchantList: []types.Merchant{
					{
						Id: 2,
					},
				},
				MerchantCount: 0,
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
