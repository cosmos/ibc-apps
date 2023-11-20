package keeper

import (
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/testing/simapp/x/dummyware/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis
func (k Keeper) InitGenesis(ctx sdk.Context, state types.GenesisState) {
	if err := k.SetParams(ctx, state.Params); err != nil {
		panic(err)
	}
}

// ExportGenesis
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	store := ctx.KVStore(k.storeKey)

	inFlightPackets := make(map[string]types.InFlightPacket)

	itr := store.Iterator(nil, nil)
	for ; itr.Valid(); itr.Next() {
		var inFlightPacket types.InFlightPacket
		k.cdc.MustUnmarshal(itr.Value(), &inFlightPacket)
		inFlightPackets[string(itr.Key())] = inFlightPacket
	}
	return &types.GenesisState{Params: k.GetParams(ctx), InFlightPackets: inFlightPackets}
}
