package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name
	ModuleName = "lojas"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// BaseDenom é a unidade mínima on-chain da BYX.
	BaseDenom = "ubyx"
	// DisplayDenom é a unidade exibida para humanos.
	DisplayDenom = "BYX"
	// DenomBYX mantém compatibilidade com o código existente, apontando para a unidade base.
	DenomBYX = BaseDenom
	// GovModuleName duplicates the gov module's name to avoid a dependency with x/gov.
	// It should be synced with the gov module's name if it is ever changed.
	// See: https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/x/gov/types/keys.go#L9
	GovModuleName = "gov"
)

// ParamsKey is the prefix to retrieve all Params
var ParamsKey = collections.NewPrefix("p_lojas")

var (
	MerchantKey              = collections.NewPrefix("merchant/value/")
	MerchantCountKey         = collections.NewPrefix("merchant/count/")
	SalesKey                 = collections.NewPrefix("sales/value/")
	SaleCountKey             = collections.NewPrefix("sales/count/")
	SalesByLojaIndexKey      = collections.NewPrefix("sales/by_loja/")
	DailyCashbackByLojaKey   = collections.NewPrefix("sales/daily_cashback/")
	SalesCountStateByLojaKey = collections.NewPrefix("sales/count_block_state/")
)

var (
	MerchantKeyPrefix       = []byte{0x01}
	MerchantByCreatorPrefix = []byte{0x02}
	NextMerchantIDKey       = []byte{0x03}
)
