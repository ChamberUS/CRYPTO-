package types

import (
	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInvalidParams = errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "feesplit invalid params")
)

// ErrInvalidParamsWrapf is a helper to wrap parameter validation errors.
func ErrInvalidParamsWrapf(format string, args ...interface{}) error {
	return errorsmod.Wrapf(ErrInvalidParams, format, args...)
}
