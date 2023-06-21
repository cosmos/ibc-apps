package v2

import (
	v1 "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v4/router/migrations/v1"
	routertypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v4/router/types"
)

// MigrateGenState accepts exported v1 packetforwardmiddleware genesis state and migrates it to
// v2 packetforwardmiddleware genesis state. The migration includes:
// Introduce inFlightPackets map for keeping track of forwarded packet state.
func MigrateGenState(oldState *v1.GenesisState) *routertypes.GenesisState {
	return &routertypes.GenesisState{
		Params:          routertypes.Params(oldState.Params),
		InFlightPackets: make(map[string]routertypes.InFlightPacket),
	}
}
