package interquery

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/ibc-apps/modules/async-icq/v7/interchain-query-demo/testutil/sample"
	interquerysimulation "github.com/cosmos/ibc-apps/modules/async-icq/v7/interchain-query-demo/x/interquery/simulation"
	"github.com/cosmos/ibc-apps/modules/async-icq/v7/interchain-query-demo/x/interquery/types"
)

// avoid unused import issue
var (
	_ = sample.AccAddress
	_ = interquerysimulation.FindAccount
	_ = simappparams.StakePerAccount
	_ = simulation.MsgEntryKind
	_ = baseapp.Paramspace
)

const (
	opWeightMsgSendQueryAllBalances = "op_weight_msg_send_query_all_balances"
	// TODO: Determine the simulation weight value
	defaultWeightMsgSendQueryAllBalances int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	interqueryGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		PortId: types.PortID,
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&interqueryGenesis)
}

// ProposalContents doesn't return any content functions for governance proposals
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized  param changes for the simulator
func (am AppModule) RandomizedParams(_ *rand.Rand) []simtypes.ParamChange {

	return []simtypes.ParamChange{}
}

// RegisterStoreDecoder registers a decoder
func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgSendQueryAllBalances int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgSendQueryAllBalances, &weightMsgSendQueryAllBalances, nil,
		func(_ *rand.Rand) {
			weightMsgSendQueryAllBalances = defaultWeightMsgSendQueryAllBalances
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgSendQueryAllBalances,
		interquerysimulation.SimulateMsgSendQueryAllBalances(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}
