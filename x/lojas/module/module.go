package lojas

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/buynnex-corp/byx/x/lojas/keeper"
	"github.com/buynnex-corp/byx/x/lojas/types"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

var (
	_ module.AppModuleBasic     = (*AppModuleBasic)(nil)
	_ appmodule.AppModule       = (*AppModule)(nil)
	_ appmodule.HasBeginBlocker = (*AppModule)(nil)
	_ appmodule.HasEndBlocker   = (*AppModule)(nil)
)

type AppModuleBasic struct{ cdc codec.Codec }

func (AppModuleBasic) Name() string                                { return types.ModuleName }
func (AppModuleBasic) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}
func (AppModuleBasic) RegisterInterfaces(reg codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
}
func (AppModuleBasic) RegisterGRPCGatewayRoutes(ctx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(ctx)); err != nil {
		panic(err)
	}
}

type AppModule struct {
	AppModuleBasic
	keeper     keeper.Keeper
	authKeeper types.AuthKeeper
	bankKeeper types.BankKeeper
}

func (AppModule) IsAppModule() {}

func NewAppModule(cdc codec.Codec, k keeper.Keeper, auth types.AuthKeeper, bank types.BankKeeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         k,
		authKeeper:     auth,
		bankKeeper:     bank,
	}
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	// Msg usa valor; Query usa PONTEIRO
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), &am.keeper)
}

/************ Genesis ************/
func (am AppModule) DefaultGenesis(codec.JSONCodec) json.RawMessage {
	return am.cdc.MustMarshalJSON(types.DefaultGenesis())
}
func (am AppModule) ValidateGenesis(_ codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var gs types.GenesisState
	if err := am.cdc.UnmarshalJSON(bz, &gs); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return gs.Validate()
}
func (am AppModule) InitGenesis(ctx sdk.Context, _ codec.JSONCodec, b json.RawMessage) {
	var gs types.GenesisState
	if err := am.cdc.UnmarshalJSON(b, &gs); err != nil {
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
	bz, err := am.cdc.MarshalJSON(gs)
	if err != nil {
		panic(err)
	}
	return bz
}
func (am AppModule) BeginBlock(context.Context) error { return nil }
func (am AppModule) EndBlock(context.Context) error   { return nil }
