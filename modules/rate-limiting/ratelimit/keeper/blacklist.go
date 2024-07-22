package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/ibc-rate-limiting/ratelimit/types"
)

// Adds a denom to a blacklist to prevent all IBC transfers with this denom
func (k Keeper) AddDenomToBlacklist(ctx sdk.Context, denom string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DenomBlacklistKeyPrefix)
	key := types.KeyPrefix(denom)
	store.Set(key, []byte{1})
}

// Removes a denom from a blacklist to re-enable IBC transfers for that denom
func (k Keeper) RemoveDenomFromBlacklist(ctx sdk.Context, denom string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DenomBlacklistKeyPrefix)
	key := types.KeyPrefix(denom)
	store.Delete(key)
}

// Check if a denom is currently blacklisted
func (k Keeper) IsDenomBlacklisted(ctx sdk.Context, denom string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DenomBlacklistKeyPrefix)

	key := types.KeyPrefix(denom)
	value := store.Get(key)
	found := len(value) != 0

	return found
}

// Get all the blacklisted denoms
func (k Keeper) GetAllBlacklistedDenoms(ctx sdk.Context) []string {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DenomBlacklistKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	allBlacklistedDenoms := []string{}
	for ; iterator.Valid(); iterator.Next() {
		allBlacklistedDenoms = append(allBlacklistedDenoms, string(iterator.Key()))
	}

	return allBlacklistedDenoms
}
