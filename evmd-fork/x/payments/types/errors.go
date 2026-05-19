package types

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

// x/payments module sentinel errors
var (
	ErrInvalidSigner            = errors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrPaymentRequestNotFound   = errors.Register(ModuleName, 1101, "payment request not found")
	ErrPaymentRequestExpired    = errors.Register(ModuleName, 1102, "payment request expired")
	ErrPaymentRequestNotPending = errors.Register(ModuleName, 1103, "payment request not pending")
)
