package keeper_test

import (
	"fmt"
	"time"

	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"

	sdkmath "cosmossdk.io/math"
)

// Store a rate limit with a non-zero flow for each duration
func (s *KeeperTestSuite) resetRateLimits(denom string, durations []uint64, nonZeroFlow int64) {
	// Add/reset rate limit with a quota duration hours for each duration in the list
	for i, duration := range durations {
		channelId := fmt.Sprintf("channel-%d", i)

		s.App.RatelimitKeeper.SetRateLimit(s.Ctx, types.RateLimit{
			Path: &types.Path{
				Denom:             denom,
				ChannelOrClientId: channelId,
			},
			Quota: &types.Quota{
				DurationHours: duration,
			},
			Flow: &types.Flow{
				Inflow:       sdkmath.NewInt(nonZeroFlow),
				Outflow:      sdkmath.NewInt(nonZeroFlow),
				ChannelValue: sdkmath.NewInt(100),
			},
		})
	}
}

func (s *KeeperTestSuite) TestBeginBlocker() {
	// We'll create three rate limits with different durations
	// And then pass in epoch ids that will cause each to trigger a reset in order
	// i.e. epochId 2   will only cause duration 2 to trigger (2 % 2 == 0; and 9 % 2 != 0; 25 % 2 != 0),
	//      epochId 9,  will only cause duration 3 to trigger (9 % 2 != 0; and 9 % 3 == 0; 25 % 3 != 0)
	//      epochId 25, will only cause duration 5 to trigger (9 % 5 != 0; and 9 % 5 != 0; 25 % 5 == 0)
	durations := []uint64{2, 3, 5}
	epochIds := []uint64{2, 9, 25}
	nonZeroFlow := int64(10)

	blockTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	s.Ctx = s.Ctx.WithBlockTime(blockTime)

	for i, epochId := range epochIds {
		// First reset the  rate limits to they have a non-zero flow
		s.resetRateLimits(denom, durations, nonZeroFlow)

		duration := durations[i]
		channelIdFromResetRateLimit := fmt.Sprintf("channel-%d", i)

		// Setup epochs so that the hook triggers
		// (epoch start time + duration must be before block time)
		s.App.RatelimitKeeper.SetHourEpoch(s.Ctx, types.HourEpoch{
			EpochNumber:    epochId - 1,
			Duration:       time.Minute,
			EpochStartTime: blockTime.Add(-2 * time.Minute),
		})
		s.App.RatelimitKeeper.BeginBlocker(s.Ctx)

		// Check rate limits (only one rate limit should reset for each hook trigger)
		rateLimits := s.App.RatelimitKeeper.GetAllRateLimits(s.Ctx)
		for _, rateLimit := range rateLimits {
			context := fmt.Sprintf("duration: %d, epoch: %d", duration, epochId)

			if rateLimit.Path.ChannelOrClientId == channelIdFromResetRateLimit {
				s.Require().Equal(int64(0), rateLimit.Flow.Inflow.Int64(), "inflow was not reset to 0 - %s", context)
				s.Require().Equal(int64(0), rateLimit.Flow.Outflow.Int64(), "outflow was not reset to 0 - %s", context)
			} else {
				s.Require().Equal(nonZeroFlow, rateLimit.Flow.Inflow.Int64(), "inflow should have been left unchanged - %s", context)
				s.Require().Equal(nonZeroFlow, rateLimit.Flow.Outflow.Int64(), "outflow should have been left unchanged - %s", context)
			}
		}
	}
}
