package types

// DONTCOVER

import "cosmossdk.io/errors"

// x/certificados module sentinel errors
var (
	ErrInvalidSigner         = errors.Register(ModuleName, 1200, "invalid signer")
	ErrCertificateNotFound   = errors.Register(ModuleName, 1201, "certificate not found")
	ErrNotMerchantOwner      = errors.Register(ModuleName, 1202, "not merchant owner")
	ErrOwnerMismatch         = errors.Register(ModuleName, 1203, "certificate owner mismatch")
	ErrCertificateRevoked    = errors.Register(ModuleName, 1204, "certificate revoked")
	ErrModuleDisabled        = errors.Register(ModuleName, 1205, "module disabled")
	ErrTransferNotAllowed    = errors.Register(ModuleName, 1206, "transfer not allowed")
	ErrInvalidServiceRequest = errors.Register(ModuleName, 1207, "invalid service request")
)
