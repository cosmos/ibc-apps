package interquery

import (
	"github.com/cosmos/ibc-apps/modules/async-icq/v8/interchain-query-demo/x/interquery/keeper"
	"github.com/cosmos/ibc-apps/modules/async-icq/v8/interchain-query-demo/x/interquery/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
//
// NOTE: ibc-go v11 no longer uses the capability module, so port binding is
// handled by the IBC router. We only persist the port id and params here.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetPort(ctx, genState.PortId)
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)
	genesis.PortId = k.GetPort(ctx)

	return genesis
}
