package feesplit

import (
	"cosmossdk.io/core/appmodule"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"github.com/cosmos/cosmos-sdk/codec"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/buynnex/iaos-evmd/x/feesplit/keeper"
	"github.com/buynnex/iaos-evmd/x/feesplit/types"
)

var _ depinject.OnePerModuleType = AppModule{}

func (AppModule) IsOnePerModuleType() {}

func init() {
	appconfig.Register(
		&types.Module{},
		appconfig.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	Config       *types.Module
	Cdc          codec.Codec
	StoreService corestore.KVStoreService
	BankKeeper   bankkeeper.Keeper
}

type ModuleOutputs struct {
	depinject.Out

	FeesplitKeeper keeper.Keeper
	Module         appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	k := keeper.NewKeeper(
		in.StoreService,
		in.Cdc,
		in.BankKeeper,
	)
	m := NewAppModule(AppModuleInputs{
		Keeper: k,
		Cdc:    in.Cdc,
	})

	return ModuleOutputs{FeesplitKeeper: k, Module: m}
}
