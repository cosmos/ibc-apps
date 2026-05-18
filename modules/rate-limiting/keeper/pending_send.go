package keeper

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Sets the sequence number of a packet that was just sent
func (k Keeper) SetPendingSendPacket(ctx sdk.Context, channelId string, sequence uint64) error {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, types.PendingSendPacketPrefix)
	key, err := types.GetPendingSendPacketKey(channelId, sequence)
	if err != nil {
		return err
	}
	store.Set(key, []byte{1})
	return nil
}

// Remove a pending packet sequence number from the store
// Used after the ack or timeout for a packet has been received
func (k Keeper) RemovePendingSendPacket(ctx sdk.Context, channelId string, sequence uint64) error {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, types.PendingSendPacketPrefix)
	key, err := types.GetPendingSendPacketKey(channelId, sequence)
	if err != nil {
		return err
	}

	store.Delete(key)
	return nil
}

// Checks whether the packet sequence number is in the store - indicating that it was
// sent during the current quota
func (k Keeper) CheckPacketSentDuringCurrentQuota(ctx sdk.Context, channelId string, sequence uint64) (bool, error) {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, types.PendingSendPacketPrefix)
	key, err := types.GetPendingSendPacketKey(channelId, sequence)
	if err != nil {
		return false, err
	}
	valueBz := store.Get(key)
	found := len(valueBz) != 0
	return found, nil
}

// Get all pending packet sequence numbers
func (k Keeper) GetAllPendingSendPackets(ctx sdk.Context) (pendingPackets []string, err error) {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, types.PendingSendPacketPrefix)

	iterator := store.Iterator(nil, nil)
	defer func() {
		err = iterator.Close()
	}()

	pendingPackets = make([]string, 0)
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()

		channelId := string(key[:types.PendingSendPacketChannelLength])
		channelId = strings.TrimRight(channelId, "\x00") // removes null bytes from suffix
		sequence := binary.BigEndian.Uint64(key[types.PendingSendPacketChannelLength:])

		packetId := fmt.Sprintf("%s/%d", channelId, sequence)
		pendingPackets = append(pendingPackets, packetId)
	}

	return pendingPackets, nil
}

// Remove all pending sequence numbers from the store
// This is executed when the quota resets
func (k Keeper) RemoveAllChannelPendingSendPackets(ctx sdk.Context, channelId string) (err error) {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, types.PendingSendPacketPrefix)

	if len(channelId) > types.PendingSendPacketChannelLength {
		return errorsmod.Wrapf(types.ErrInvalidChannelId, "channel %s with length %d is greater than the allowed length %d", channelId, len(channelId), types.PendingSendPacketChannelLength)
	}

	channelIDBz := make([]byte, types.PendingSendPacketChannelLength)
	copy(channelIDBz, channelId)

	iterator := storetypes.KVStorePrefixIterator(store, channelIDBz)
	defer func() {
		err = errors.Join(err, iterator.Close())
	}()

	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}
	return nil
}
