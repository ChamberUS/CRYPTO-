package feesplit

import (
	"context"
	"encoding/json"

	"github.com/buynnex/iaos-evmd/x/feesplit/keeper"
	"github.com/buynnex/iaos-evmd/x/feesplit/types"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

var (
	_ module.AppModuleBasic     = (*AppModuleBasic)(nil)
	_ appmodule.AppModule       = (*AppModule)(nil)
	_ appmodule.HasBeginBlocker = (*AppModule)(nil)
)

type AppModuleBasic struct{ cdc codec.Codec }

func (AppModuleBasic) Name() string                                { return types.ModuleName }
func (AppModuleBasic) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}
func (AppModuleBasic) RegisterInterfaces(reg codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
}
func (AppModuleBasic) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {}

func (AppModule) Type() string {
	return "byx.feesplit.v1"
}

type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

func (AppModule) IsAppModule() {}

type AppModuleInputs struct {
	depinject.In

	Keeper keeper.Keeper
	Cdc    codec.Codec
}

func NewAppModule(in AppModuleInputs) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: in.Cdc},
		keeper:         in.Keeper,
	}
}

func (am AppModule) RegisterServices(registrar grpc.ServiceRegistrar) error {
	// no msg/query services yet
	_ = registrar
	return nil
}

func (am AppModule) DefaultGenesis(codec.JSONCodec) json.RawMessage {
	bz, err := json.Marshal(types.DefaultGenesis())
	if err != nil {
		panic(err)
	}
	return bz
}

func (am AppModule) ValidateGenesis(_ codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var gs types.GenesisState
	if err := json.Unmarshal(bz, &gs); err != nil {
		return err
	}
	return gs.Validate()
}

func (am AppModule) InitGenesis(ctx sdk.Context, _ codec.JSONCodec, bz json.RawMessage) {
	var gs types.GenesisState
	if err := json.Unmarshal(bz, &gs); err != nil {
		panic(err)
	}
	if err := am.keeper.InitGenesis(ctx, gs); err != nil {
		panic(err)
	}
}

func (am AppModule) ExportGenesis(ctx sdk.Context, _ codec.JSONCodec) json.RawMessage {
	gs, err := am.keeper.ExportGenesis(ctx)
	if err != nil {
		panic(err)
	}
	bz, err := json.Marshal(gs)
	if err != nil {
		panic(err)
	}
	return bz
}

func (AppModule) ConsensusVersion() uint64 { return 1 }

func (am AppModule) BeginBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return am.keeper.BeginBlocker(sdkCtx)
}
