package types

import (
	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:              DefaultParams(),
		PaymentRequests:     []PaymentRequest{},
		PaymentRequestCount: 0,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	seen := make(map[uint64]struct{})
	for _, pr := range gs.PaymentRequests {
		if _, exists := seen[pr.Id]; exists {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "duplicate payment request id %d", pr.Id)
		}
		seen[pr.Id] = struct{}{}
	}

	return nil
}
