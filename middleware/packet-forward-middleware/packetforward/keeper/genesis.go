package keeper

import (
<<<<<<< HEAD:middleware/packet-forward-middleware/router/keeper/genesis.go
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v5/router/types"
=======
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
>>>>>>> 47f2ae0 (rename: `router` -> `packetforward` (#118)):middleware/packet-forward-middleware/packetforward/keeper/genesis.go

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis
func (k Keeper) InitGenesis(ctx sdk.Context, state types.GenesisState) {
	k.SetParams(ctx, state.Params)

	// Initialize store refund path for forwarded packets in genesis state that have not yet been acked.
	store := ctx.KVStore(k.storeKey)
	for key, value := range state.InFlightPackets {
		key := key
		value := value
		bz := k.cdc.MustMarshal(&value)
		store.Set([]byte(key), bz)
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
