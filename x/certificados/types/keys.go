package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name.
	ModuleName = "certificados"
	// StoreKey defines the primary module store key.
	StoreKey = ModuleName
	// RouterKey defines the module routing key.
	RouterKey = ModuleName

	// GovModuleName duplicates the gov module name to avoid a dependency on x/gov.
	GovModuleName = "gov"
)

var (
	ParamsKey               = collections.NewPrefix("p_certificados")
	CertificateKey          = collections.NewPrefix("certificates/value/")
	CertificateCountKey     = collections.NewPrefix("certificates/count/")
	CertificateByOwnerIndex = collections.NewPrefix("certificates/by_owner/")
	CertificateByMerchant   = collections.NewPrefix("certificates/by_merchant/")
	CertificateBySerial     = collections.NewPrefix("certificates/by_serial/")
	ServiceRecordKey        = collections.NewPrefix("service_records/value/")
	ServiceCountKey         = collections.NewPrefix("service_records/count/")
)
