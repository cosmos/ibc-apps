package keeper

import (
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis
func (k Keeper) InitGenesis(ctx sdk.Context, state types.GenesisState) {
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
	return &types.GenesisState{InFlightPackets: inFlightPackets}
}
