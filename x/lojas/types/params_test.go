package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParamsValidateRanges(t *testing.T) {
	cases := []struct {
		name   string
		params Params
		expErr bool
	}{
		{
			name:   "default valid",
			params: DefaultParams(),
		},
		{
			name: "faucet max empty allowed",
			params: Params{
				FaucetEnabled:                   true,
				FaucetMaxPerTx:                  "",
				CashbackRateMicroByxPerReal:     100,
				MaxValorVendaEmCentavos:         1_000,
				MaxCashbackMicroByxPorVenda:     1_000,
				MaxCashbackDailyPerLojaMicrobyx: 1_000,
				MaxSalesPerBlockPerLoja:         1,
				CashbackDailyRetentionDays:      90,
			},
		},
		{
			name: "cashback too high",
			params: Params{
				FaucetEnabled:               true,
				FaucetMaxPerTx:              "10",
				CashbackRateMicroByxPerReal: 2_000_000,
				MaxValorVendaEmCentavos:     1_000,
				MaxCashbackMicroByxPorVenda: 1_000,
			},
			expErr: true,
		},
		{
			name: "max valor zero",
			params: Params{
				FaucetEnabled:               true,
				FaucetMaxPerTx:              "10",
				CashbackRateMicroByxPerReal: 100,
				MaxValorVendaEmCentavos:     0,
				MaxCashbackMicroByxPorVenda: 1_000,
			},
			expErr: true,
		},
		{
			name: "max cashback too high",
			params: Params{
				FaucetEnabled:               true,
				FaucetMaxPerTx:              "10",
				CashbackRateMicroByxPerReal: 100,
				MaxValorVendaEmCentavos:     1_000,
				MaxCashbackMicroByxPorVenda: 2_000_000,
			},
			expErr: true,
		},
		{
			name: "invalid faucet admin",
			params: Params{
				FaucetEnabled:               true,
				FaucetMaxPerTx:              "10",
				CashbackRateMicroByxPerReal: 100,
				MaxValorVendaEmCentavos:     1_000,
				MaxCashbackMicroByxPorVenda: 1_000,
				FaucetAdmin:                 "not-bech32",
			},
			expErr: true,
		},
		{
			name: "invalid retention",
			params: Params{
				FaucetEnabled:                   true,
				FaucetMaxPerTx:                  "10",
				CashbackRateMicroByxPerReal:     100,
				MaxValorVendaEmCentavos:         1_000,
				MaxCashbackMicroByxPorVenda:     1_000,
				MaxCashbackDailyPerLojaMicrobyx: 1_000,
				MaxSalesPerBlockPerLoja:         1,
				CashbackDailyRetentionDays:      0,
			},
			expErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
