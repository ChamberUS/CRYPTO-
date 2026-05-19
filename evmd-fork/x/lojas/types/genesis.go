package types

import "fmt"

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:       DefaultParams(),
		MerchantList: []Merchant{}}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	merchantIdMap := make(map[uint64]bool)
	merchantCount := gs.GetMerchantCount()
	for _, elem := range gs.MerchantList {
		if _, ok := merchantIdMap[elem.Id]; ok {
			return fmt.Errorf("duplicated id for merchant")
		}
		if elem.Id >= merchantCount {
			return fmt.Errorf("merchant id should be lower or equal than the last id")
		}
		merchantIdMap[elem.Id] = true
	}

	if gs.Params == (Params{}) {
		gs.Params = DefaultParams()
	}

	return gs.Params.Validate()
}
