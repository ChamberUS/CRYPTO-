package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name
	ModuleName = "payments"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module routing key
	RouterKey = ModuleName

	// GovModuleName duplicates the gov module's name to avoid a dependency with x/gov.
	// It should be synced with the gov module's name if it is ever changed.
	// See: https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/x/gov/types/keys.go#L9
	GovModuleName = "gov"
)

// ParamsKey is the prefix to retrieve all Params
var ParamsKey = collections.NewPrefix("p_payments")

var (
	PaymentRequestKey       = collections.NewPrefix("payment_request/value/")
	PaymentRequestCountKey  = collections.NewPrefix("payment_request/count/")
	PaymentRequestsByLojaPK = collections.NewPrefix("payment_request/by_loja/")
)

var (
	PaymentRequestKeyPrefix    = []byte{0x01}
	PaymentRequestByLojaPrefix = []byte{0x02}
	NextPaymentRequestIDKey    = []byte{0x03}
	PaymentRequestDedupePrefix = []byte{0x04}
)
