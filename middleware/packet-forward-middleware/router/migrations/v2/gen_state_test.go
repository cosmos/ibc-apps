package v2_test

import (
	"testing"

	v1 "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v4/router/migrations/v1"
	v2 "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v4/router/migrations/v2"
	routertypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v4/router/types"
	"github.com/stretchr/testify/require"
)

func TestMigrateGenState(t *testing.T) {
	tests := []struct {
		name     string
		oldState *v1.GenesisState
		newState *routertypes.GenesisState
	}{
		{
			name:     "successful migration introducing inFlightPackets",
			oldState: &v1.GenesisState{Params: v1.Params{FeePercentage: routertypes.DefaultFeePercentage}},
			newState: &routertypes.GenesisState{
				Params:          routertypes.Params{FeePercentage: routertypes.DefaultFeePercentage},
				InFlightPackets: make(map[string]routertypes.InFlightPacket),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actualState := v2.MigrateGenState(tc.oldState)
			require.Equal(t, tc.newState, actualState)
			require.NoError(t, actualState.Validate())
		})
	}
}
