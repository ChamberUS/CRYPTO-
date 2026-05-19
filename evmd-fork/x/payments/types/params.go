package types

import (
	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewParams creates a new Params instance.
func NewParams(defaultExp, minExp, maxExp uint64) Params {
	return Params{
		DefaultExpiresInSeconds: defaultExp,
		MinExpiresInSeconds:     minExp,
		MaxExpiresInSeconds:     maxExp,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(600, 60, 86400)
}

// Validate validates the set of params.
func (p Params) Validate() error {
	if p.MinExpiresInSeconds == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "min_expires_in_seconds must be > 0")
	}
	if p.MaxExpiresInSeconds == 0 || p.MaxExpiresInSeconds < p.MinExpiresInSeconds {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "max_expires_in_seconds must be >= min_expires_in_seconds")
	}
	if p.DefaultExpiresInSeconds < p.MinExpiresInSeconds || p.DefaultExpiresInSeconds > p.MaxExpiresInSeconds {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "default_expires_in_seconds must be within min/max range")
	}

	return nil
}
