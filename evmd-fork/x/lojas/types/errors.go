package types

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

// x/lojas module sentinel errors
var (
	ErrInvalidSigner      = errors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrDailyLimitExceeded = errors.Register(ModuleName, 1101, "daily cashback limit exceeded")
	ErrBlockLimitExceeded = errors.Register(ModuleName, 1102, "block sales limit exceeded")
)
