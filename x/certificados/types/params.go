package types

import (
	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// DefaultIssueFeeByx is the fixed issuance fee in byx (1 byx = R$0.01).
	DefaultIssueFeeByx uint64 = 1499
)

// NewParams creates a new Params instance.
func NewParams(enabled bool, issueFeeByx uint64, allowTransfer bool) Params {
	return Params{
		Enabled:       enabled,
		IssueFeeByx:   issueFeeByx,
		AllowTransfer: allowTransfer,
	}
}

// DefaultParams returns the default module parameters.
func DefaultParams() Params {
	return NewParams(true, DefaultIssueFeeByx, true)
}

// Validate validates the set of params.
func (p Params) Validate() error {
	if p.IssueFeeByx == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "issue_fee_byx must be > 0")
	}
	return nil
}
