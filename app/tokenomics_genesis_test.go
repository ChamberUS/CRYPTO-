package app_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/buynnex-corp/byx/x/lojas/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisTokenomicsSupplyCapAndMintDisabled(t *testing.T) {
	content, err := os.ReadFile("../genesis.json")
	require.NoError(t, err)

	var genesis struct {
		AppState struct {
			Bank struct {
				Supply []struct {
					Denom  string `json:"denom"`
					Amount string `json:"amount"`
				} `json:"supply"`
			} `json:"bank"`
			Mint struct {
				Minter struct {
					Inflation string `json:"inflation"`
				} `json:"minter"`
				Params struct {
					MintDenom           string `json:"mint_denom"`
					InflationRateChange string `json:"inflation_rate_change"`
					InflationMax        string `json:"inflation_max"`
					InflationMin        string `json:"inflation_min"`
				} `json:"params"`
			} `json:"mint"`
		} `json:"app_state"`
	}
	require.NoError(t, json.Unmarshal(content, &genesis))

	var byxSupply string
	for _, coin := range genesis.AppState.Bank.Supply {
		if coin.Denom == types.DenomBYX {
			byxSupply = coin.Amount
			break
		}
	}
	require.Equal(t, "1000000000000000", byxSupply)
	require.Equal(t, types.BaseDenom, genesis.AppState.Mint.Params.MintDenom)
	require.Equal(t, "0.000000000000000000", genesis.AppState.Mint.Minter.Inflation)
	require.Equal(t, "0.000000000000000000", genesis.AppState.Mint.Params.InflationRateChange)
	require.Equal(t, "0.000000000000000000", genesis.AppState.Mint.Params.InflationMax)
	require.Equal(t, "0.000000000000000000", genesis.AppState.Mint.Params.InflationMin)
}
