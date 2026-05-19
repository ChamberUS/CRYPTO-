package lojas

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/buynnex-corp/byx/testutil/sample"
	lojassimulation "github.com/buynnex-corp/byx/x/lojas/simulation"
	"github.com/buynnex-corp/byx/x/lojas/types"
)

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	lojasGenesis := types.GenesisState{
		Params:       types.DefaultParams(),
		MerchantList: []types.Merchant{{Id: 1, Creator: sample.AccAddress()}, {Id: 2, Creator: sample.AccAddress()}}, MerchantCount: 3,
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&lojasGenesis)
}

// RegisterStoreDecoder registers a decoder.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)
	const (
		opWeightMsgCreateLojista          = "op_weight_msg_lojas"
		defaultWeightMsgCreateLojista int = 100
	)

	var weightMsgCreateLojista int
	simState.AppParams.GetOrGenerate(opWeightMsgCreateLojista, &weightMsgCreateLojista, nil,
		func(_ *rand.Rand) {
			weightMsgCreateLojista = defaultWeightMsgCreateLojista
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreateLojista,
		lojassimulation.SimulateMsgCreateLojista(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgTransferirByx          = "op_weight_msg_lojas"
		defaultWeightMsgTransferirByx int = 100
	)

	var weightMsgTransferirByx int
	simState.AppParams.GetOrGenerate(opWeightMsgTransferirByx, &weightMsgTransferirByx, nil,
		func(_ *rand.Rand) {
			weightMsgTransferirByx = defaultWeightMsgTransferirByx
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgTransferirByx,
		lojassimulation.SimulateMsgTransferirByx(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgCreateMerchant          = "op_weight_msg_lojas"
		defaultWeightMsgCreateMerchant int = 100
	)

	var weightMsgCreateMerchant int
	simState.AppParams.GetOrGenerate(opWeightMsgCreateMerchant, &weightMsgCreateMerchant, nil,
		func(_ *rand.Rand) {
			weightMsgCreateMerchant = defaultWeightMsgCreateMerchant
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreateMerchant,
		lojassimulation.SimulateMsgCreateMerchant(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgUpdateMerchant          = "op_weight_msg_lojas"
		defaultWeightMsgUpdateMerchant int = 100
	)

	var weightMsgUpdateMerchant int
	simState.AppParams.GetOrGenerate(opWeightMsgUpdateMerchant, &weightMsgUpdateMerchant, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateMerchant = defaultWeightMsgUpdateMerchant
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateMerchant,
		lojassimulation.SimulateMsgUpdateMerchant(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgDeleteMerchant          = "op_weight_msg_lojas"
		defaultWeightMsgDeleteMerchant int = 100
	)

	var weightMsgDeleteMerchant int
	simState.AppParams.GetOrGenerate(opWeightMsgDeleteMerchant, &weightMsgDeleteMerchant, nil,
		func(_ *rand.Rand) {
			weightMsgDeleteMerchant = defaultWeightMsgDeleteMerchant
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgDeleteMerchant,
		lojassimulation.SimulateMsgDeleteMerchant(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{}
}
