package types

import (
	"context"

	certificadostypes "github.com/buynnex-corp/byx/x/certificados/types"
	lojastypes "github.com/buynnex-corp/byx/x/lojas/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected interface for the Bank module.
type BankKeeper interface {
	SendCoins(context.Context, sdk.AccAddress, sdk.AccAddress, sdk.Coins) error
}

// LojasKeeper defines the expected subset we need from x/lojas.
type LojasKeeper interface {
	GetMerchant(context.Context, uint64) (lojastypes.Merchant, error)
}

// CertificadosKeeper defines the expected subset we need from x/certificados.
type CertificadosKeeper interface {
	GetCertificate(context.Context, uint64) (certificadostypes.Certificate, error)
	TransferCertificate(context.Context, uint64, string, string) error
	GetParams(context.Context) (certificadostypes.Params, error)
}
