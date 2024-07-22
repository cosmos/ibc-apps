package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/ibc-rate-limiting/ratelimit/types"
)

// Stores the hour epoch
func (k Keeper) SetHourEpoch(ctx sdk.Context, epoch types.HourEpoch) {
	store := ctx.KVStore(k.storeKey)
	epochBz := k.cdc.MustMarshal(&epoch)
	store.Set(types.HourEpochKey, epochBz)
}

// Reads the hour epoch from the store
func (k Keeper) GetHourEpoch(ctx sdk.Context) (epoch types.HourEpoch) {
	store := ctx.KVStore(k.storeKey)

	epochBz := store.Get(types.HourEpochKey)
	if len(epochBz) == 0 {
		panic("Hour epoch not found")
	}

	k.cdc.MustUnmarshal(epochBz, &epoch)
	return epoch
}

// Checks if it's time to start the new hour epoch
func (k Keeper) CheckHourEpochStarting(ctx sdk.Context) (epochStarting bool, epochNumber uint64) {
	hourEpoch := k.GetHourEpoch(ctx)

	// If the block time is later than the current epoch start time + epoch duration,
	// move onto the next epoch by incrementing the epoch number, height, and start time
	currentEpochEndTime := hourEpoch.EpochStartTime.Add(hourEpoch.Duration)
	shouldNextEpochStart := ctx.BlockTime().After(currentEpochEndTime)
	if shouldNextEpochStart {
		hourEpoch.EpochNumber++
		hourEpoch.EpochStartTime = currentEpochEndTime
		hourEpoch.EpochStartHeight = ctx.BlockHeight()

		k.SetHourEpoch(ctx, hourEpoch)
		return true, hourEpoch.EpochNumber
	}

	// Otherwise, indicate that a new epoch is not starting
	return false, 0
}
