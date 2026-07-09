package keeper

import (
	"fmt"

	"github.com/cosmos/ibc-apps/modules/async-icq/v8/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the icq state and sets the PortID.
func (k Keeper) InitGenesis(ctx sdk.Context, state types.GenesisState) {
	k.SetPort(ctx, state.HostPort)

	if err := k.SetParams(ctx, state.Params); err != nil {
		panic(fmt.Sprintf("could not set params: %v", err))
	}
}

// ExportGenesis exports icq module's portID and denom trace info into its genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		HostPort: k.GetPort(ctx),
		Params:   k.GetParams(ctx),
	}
}
