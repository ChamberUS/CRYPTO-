package lojas

import (
	corestore "cosmossdk.io/core/store"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"github.com/cosmos/cosmos-sdk/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/buynnex/iaos-evmd/x/lojas/keeper"
	"github.com/buynnex/iaos-evmd/x/lojas/types"
)

var _ depinject.OnePerModuleType = AppModule{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (AppModule) IsOnePerModuleType() {}

func init() {
	appmodule.Register(
		&types.Module{},
		appmodule.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	Config       *types.Module
	StoreService corestore.KVStoreService
	Cdc          codec.Codec
	AddressCodec address.Codec

	AccountKeeper authkeeper.AccountKeeper
	BankKeeper    bankkeeper.Keeper
}

type ModuleOutputs struct {
	depinject.Out

	Keeper keeper.Keeper
	Module appmodule.AppModule
}

func ProvideModule(in ModuleInputs) (ModuleOutputs, error) {
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	if in.Config != nil && in.Config.Authority != "" {
		authority = authtypes.NewModuleAddressOrBech32Address(in.Config.Authority)
	}

	k := keeper.NewKeeper(
		in.StoreService,
		in.Cdc,
		in.AddressCodec,
		authority.Bytes(),
		in.AccountKeeper,
		in.BankKeeper,
	)

	m := NewAppModule(in.Cdc, k, in.AccountKeeper, in.BankKeeper)
	return ModuleOutputs{
		Keeper: k,
		Module: m,
	}, nil
}
