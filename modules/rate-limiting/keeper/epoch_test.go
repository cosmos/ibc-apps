package keeper_test

import (
	"time"

	"github.com/cosmos/ibc-apps/modules/rate-limiting/v9/types"
)

// Tests Get/Set Hour epoch
func (s *KeeperTestSuite) TestHourEpoch() {
	expectedHourEpoch := types.HourEpoch{
		Duration:         time.Hour,
		EpochNumber:      1,
		EpochStartTime:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EpochStartHeight: 10,
	}
	s.App.RatelimitKeeper.SetHourEpoch(s.Ctx, expectedHourEpoch)

	actualHourEpoch := s.App.RatelimitKeeper.GetHourEpoch(s.Ctx)
	s.Require().Equal(expectedHourEpoch, actualHourEpoch, "hour epoch")
}

func (s *KeeperTestSuite) TestCheckHourEpochStarting() {
	epochStartTime := time.Date(2024, 1, 1, 0, 0, 1, 0, time.UTC)
	blockHeight := int64(10)
	duration := time.Minute

	initialEpoch := types.HourEpoch{
		EpochNumber:    10,
		EpochStartTime: epochStartTime,
		Duration:       duration,
	}
	nextEpoch := types.HourEpoch{
		EpochNumber:      initialEpoch.EpochNumber + 1, // epoch number increments
		EpochStartTime:   epochStartTime.Add(duration), // start time increments by duration
		EpochStartHeight: blockHeight,                  // height gets current block height
		Duration:         duration,
	}

	testCases := []struct {
		name                  string
		blockTime             time.Time
		expectedEpochStarting bool
	}{
		{
			name:                  "in middle of epoch",
			blockTime:             epochStartTime.Add(duration / 2), // halfway through epoch
			expectedEpochStarting: false,
		},
		{
			name:                  "right before epoch boundary",
			blockTime:             epochStartTime.Add(duration).Add(-1 * time.Second), // 1 second before epoch
			expectedEpochStarting: false,
		},
		{
			name:                  "at epoch boundary",
			blockTime:             epochStartTime.Add(duration), // at epoch boundary
			expectedEpochStarting: false,
		},
		{
			name:                  "right after epoch boundary",
			blockTime:             epochStartTime.Add(duration).Add(time.Second), // one second after epoch boundary
			expectedEpochStarting: true,
		},
		{
			name:                  "in middle of next epoch",
			blockTime:             epochStartTime.Add(duration).Add(duration / 2), // halfway through next epoch
			expectedEpochStarting: true,
		},
		{
			name:                  "next epoch skipped",
			blockTime:             epochStartTime.Add(duration * 10), // way after next epoch (still increments only once)
			expectedEpochStarting: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Ctx = s.Ctx.WithBlockTime(tc.blockTime)
			s.Ctx = s.Ctx.WithBlockHeight(blockHeight)

			s.App.RatelimitKeeper.SetHourEpoch(s.Ctx, initialEpoch)

			actualStarting, actualEpochNumber := s.App.RatelimitKeeper.CheckHourEpochStarting(s.Ctx)
			s.Require().Equal(tc.expectedEpochStarting, actualStarting, "epoch starting")

			expectedEpoch := initialEpoch
			if tc.expectedEpochStarting {
				expectedEpoch = nextEpoch
				s.Require().Equal(expectedEpoch.EpochNumber, actualEpochNumber, "epoch number")
			}

			actualHourEpoch := s.App.RatelimitKeeper.GetHourEpoch(s.Ctx)
			s.Require().Equal(expectedEpoch, actualHourEpoch, "hour epoch")
		})
	}
}
