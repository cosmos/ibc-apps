package keeper

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ibc-apps/modules/rate-limiting/v7/types"
)

// Sets the sequence number of a packet that was just sent
func (k Keeper) SetPendingSendPacket(ctx sdk.Context, channelId string, sequence uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PendingSendPacketPrefix)
	key := types.GetPendingSendPacketKey(channelId, sequence)
	store.Set(key, []byte{1})
}

// Remove a pending packet sequence number from the store
// Used after the ack or timeout for a packet has been received
func (k Keeper) RemovePendingSendPacket(ctx sdk.Context, channelId string, sequence uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PendingSendPacketPrefix)
	key := types.GetPendingSendPacketKey(channelId, sequence)
	store.Delete(key)
}

// Checks whether the packet sequence number is in the store - indicating that it was
// sent during the current quota
func (k Keeper) CheckPacketSentDuringCurrentQuota(ctx sdk.Context, channelId string, sequence uint64) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PendingSendPacketPrefix)
	key := types.GetPendingSendPacketKey(channelId, sequence)
	valueBz := store.Get(key)
	found := len(valueBz) != 0
	return found
}

// Get all pending packet sequence numbers
func (k Keeper) GetAllPendingSendPackets(ctx sdk.Context) []string {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PendingSendPacketPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	pendingPackets := []string{}
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()

		channelId := string(key[:types.PendingSendPacketChannelLength])
		channelId = strings.TrimRight(channelId, "\x00") // removes null bytes from suffix
		sequence := binary.BigEndian.Uint64(key[types.PendingSendPacketChannelLength:])

		packetId := fmt.Sprintf("%s/%d", channelId, sequence)
		pendingPackets = append(pendingPackets, packetId)
	}

	return pendingPackets
}

// Remove all pending sequence numbers from the store
// This is executed when the quota resets
func (k Keeper) RemoveAllChannelPendingSendPackets(ctx sdk.Context, channelId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PendingSendPacketPrefix)

	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefix(channelId))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}
}
