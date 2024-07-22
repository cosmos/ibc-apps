package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/ibc-rate-limiting/ratelimit/types"
)

// Adds an pair of sender and receiver addresses to the whitelist to allow all
// IBC transfers between those addresses to skip all flow calculations
func (k Keeper) SetWhitelistedAddressPair(ctx sdk.Context, whitelist types.WhitelistedAddressPair) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AddressWhitelistKeyPrefix)
	key := types.GetAddressWhitelistKey(whitelist.Sender, whitelist.Receiver)
	value := k.cdc.MustMarshal(&whitelist)
	store.Set(key, value)
}

// Removes a whitelisted address pair so that it's transfers are counted in the quota
func (k Keeper) RemoveWhitelistedAddressPair(ctx sdk.Context, sender, receiver string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AddressWhitelistKeyPrefix)
	key := types.GetAddressWhitelistKey(sender, receiver)
	store.Delete(key)
}

// Check if a sender/receiver address pair is currently whitelisted
func (k Keeper) IsAddressPairWhitelisted(ctx sdk.Context, sender, receiver string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AddressWhitelistKeyPrefix)

	key := types.GetAddressWhitelistKey(sender, receiver)
	value := store.Get(key)
	found := len(value) != 0

	return found
}

// Get all the whitelisted addresses
func (k Keeper) GetAllWhitelistedAddressPairs(ctx sdk.Context) []types.WhitelistedAddressPair {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AddressWhitelistKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	allWhitelistedAddresses := []types.WhitelistedAddressPair{}
	for ; iterator.Valid(); iterator.Next() {
		whitelist := types.WhitelistedAddressPair{}
		k.cdc.MustUnmarshal(iterator.Value(), &whitelist)
		allWhitelistedAddresses = append(allWhitelistedAddresses, whitelist)
	}

	return allWhitelistedAddresses
}
