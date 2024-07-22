package keeper_test

import (
	"strconv"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/ibc-apps/modules/rate-limiting/v7/types"
)

func createRateLimits() []types.RateLimit {
	rateLimits := []types.RateLimit{}
	for i := int64(1); i <= 3; i++ {
		suffix := strconv.Itoa(int(i))
		rateLimit := types.RateLimit{
			Path:  &types.Path{Denom: "denom-" + suffix, ChannelId: "channel-" + suffix},
			Quota: &types.Quota{MaxPercentSend: sdkmath.NewInt(i), MaxPercentRecv: sdkmath.NewInt(i), DurationHours: uint64(i)},
			Flow:  &types.Flow{Inflow: sdkmath.NewInt(i), Outflow: sdkmath.NewInt(i), ChannelValue: sdkmath.NewInt(i)},
		}

		rateLimits = append(rateLimits, rateLimit)
	}
	return rateLimits
}

func (s *KeeperTestSuite) TestGenesis() {
	currentHour := 13
	blockTime := time.Date(2024, 1, 1, currentHour, 55, 8, 0, time.UTC)            // 13:55:08
	defaultEpochStartTime := time.Date(2024, 1, 1, currentHour, 0, 0, 0, time.UTC) // 13:00:00 (truncated to hour)
	blockHeight := int64(10)

	testCases := []struct {
		name          string
		genesisState  types.GenesisState
		firstEpoch    bool
		expectedError string
	}{
		{
			name:         "valid default state",
			genesisState: *types.DefaultGenesis(),
			firstEpoch:   true,
		},
		{
			name: "valid custom state",
			genesisState: types.GenesisState{
				RateLimits: createRateLimits(),
				WhitelistedAddressPairs: []types.WhitelistedAddressPair{
					{Sender: "senderA", Receiver: "receiverA"},
					{Sender: "senderB", Receiver: "receiverB"},
				},
				BlacklistedDenoms:                []string{"denomA", "denomB"},
				PendingSendPacketSequenceNumbers: []string{"channel-0/1", "channel-2/3"},
				HourEpoch: types.HourEpoch{
					EpochNumber:      1,
					EpochStartTime:   blockTime,
					Duration:         time.Minute,
					EpochStartHeight: 1,
				},
			},
			firstEpoch: false,
		},
		{
			name: "invalid packet sequence - wrong delimiter",
			genesisState: types.GenesisState{
				RateLimits:                       createRateLimits(),
				PendingSendPacketSequenceNumbers: []string{"channel-0/1", "channel-2|3"},
			},
			expectedError: "invalid pending send packet (channel-2|3), must be of form: {channelId}/{sequenceNumber}",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Ctx = s.Ctx.WithBlockTime(blockTime)
			s.Ctx = s.Ctx.WithBlockHeight(blockHeight)

			// Call initGenesis with a panic wrapper for the error cases
			defer func() {
				if recoveryError := recover(); recoveryError != nil {
					s.Require().Equal(tc.expectedError, recoveryError, "expected error from panic")
				}
			}()
			s.App.RatelimitKeeper.InitGenesis(s.Ctx, tc.genesisState)

			// If the hour epoch was not uninitialized in the raw genState,
			// it will be initialized during InitGenesis
			expectedGenesis := tc.genesisState
			if tc.firstEpoch {
				expectedGenesis.HourEpoch.EpochNumber = uint64(currentHour)
				expectedGenesis.HourEpoch.EpochStartTime = defaultEpochStartTime
				expectedGenesis.HourEpoch.EpochStartHeight = blockHeight
			}

			// Check that the exported state matches the imported state
			exportedState := s.App.RatelimitKeeper.ExportGenesis(s.Ctx)
			s.Require().Equal(expectedGenesis, *exportedState, "exported genesis state")
		})
	}
}
