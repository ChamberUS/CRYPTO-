package types

// Params define configuration for fee split.
type Params struct {
	Enabled            bool     `json:"enabled" yaml:"enabled"`
	SplitBpsValidators uint32   `json:"split_bps_validators" yaml:"split_bps_validators"`
	SplitBpsTreasury   uint32   `json:"split_bps_treasury" yaml:"split_bps_treasury"`
	SplitBpsBurn       uint32   `json:"split_bps_burn" yaml:"split_bps_burn"`
	DenomsAllowlist    []string `json:"denoms_allowlist" yaml:"denoms_allowlist"`
}

func DefaultParams() Params {
	return Params{
		Enabled:            true,
		SplitBpsValidators: 6000,
		SplitBpsTreasury:   3000,
		SplitBpsBurn:       1000,
		DenomsAllowlist:    []string{DefaultDenom},
	}
}

func (p Params) Validate() error {
	total := p.SplitBpsValidators + p.SplitBpsTreasury + p.SplitBpsBurn
	if total != 10000 {
		return ErrInvalidParamsWrapf("split bps must total 10000, got %d", total)
	}
	if p.SplitBpsValidators > 10000 || p.SplitBpsTreasury > 10000 || p.SplitBpsBurn > 10000 {
		return ErrInvalidParamsWrapf("bps must be <= 10000")
	}
	if p.SplitBpsValidators == 0 && p.SplitBpsTreasury == 0 && p.SplitBpsBurn == 0 {
		return ErrInvalidParamsWrapf("at least one split bps must be positive")
	}
	// allowlist: if empty, apply to all denoms (documented). Otherwise ensure non-empty entries.
	for _, d := range p.DenomsAllowlist {
		if d == "" {
			return ErrInvalidParamsWrapf("denom allowlist cannot contain empty entries")
		}
	}
	return nil
}

// GenesisState defines the feesplit module genesis state.
type GenesisState struct {
	Params Params `json:"params" yaml:"params"`
}

func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

func (gs GenesisState) Validate() error {
	return gs.Params.Validate()
}
