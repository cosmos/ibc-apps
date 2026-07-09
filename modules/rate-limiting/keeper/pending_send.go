package keeper

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/store/v2/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Sets the sequence number of a packet that was just sent
func (k Keeper) SetPendingSendPacket(ctx sdk.Context, channelId string, sequence uint64) error {
	return k.setPendingPacket(ctx, types.PendingSendPacketPrefix, channelId, sequence)
}

// Sets the sequence number of a packet that was just received
func (k Keeper) SetPendingReceivePacket(ctx sdk.Context, channelId string, sequence uint64) error {
	return k.setPendingPacket(ctx, types.PendingReceivePacketPrefix, channelId, sequence)
}

func (k Keeper) setPendingPacket(ctx sdk.Context, keyPrefix []byte, channelId string, sequence uint64) error {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, keyPrefix)
	key, err := types.GetPendingPacketKey(channelId, sequence)
	if err != nil {
		return err
	}
	store.Set(key, []byte{1})
	return nil
}

// Remove a pending packet sequence number from the store
// Used after the ack or timeout for a packet has been received
func (k Keeper) RemovePendingSendPacket(ctx sdk.Context, channelId string, sequence uint64) error {
	return k.removePendingPacket(ctx, types.PendingSendPacketPrefix, channelId, sequence)
}

// Remove a pending receive packet sequence number from the store
// Used after an async acknowledgement has been written
func (k Keeper) RemovePendingReceivePacket(ctx sdk.Context, channelId string, sequence uint64) error {
	return k.removePendingPacket(ctx, types.PendingReceivePacketPrefix, channelId, sequence)
}

func (k Keeper) removePendingPacket(ctx sdk.Context, keyPrefix []byte, channelId string, sequence uint64) error {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, keyPrefix)
	key, err := types.GetPendingPacketKey(channelId, sequence)
	if err != nil {
		return err
	}

	store.Delete(key)
	return nil
}

// Checks whether the packet sequence number is in the store - indicating that it was
// sent during the current quota
func (k Keeper) CheckPacketSentDuringCurrentQuota(ctx sdk.Context, channelId string, sequence uint64) (bool, error) {
	return k.checkPacketDuringCurrentQuota(ctx, types.PendingSendPacketPrefix, channelId, sequence)
}

// Checks whether the packet sequence number is in the store - indicating that it was
// received during the current quota
func (k Keeper) CheckPacketReceivedDuringCurrentQuota(ctx sdk.Context, channelId string, sequence uint64) (bool, error) {
	return k.checkPacketDuringCurrentQuota(ctx, types.PendingReceivePacketPrefix, channelId, sequence)
}

func (k Keeper) checkPacketDuringCurrentQuota(ctx sdk.Context, keyPrefix []byte, channelId string, sequence uint64) (bool, error) {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, keyPrefix)
	key, err := types.GetPendingPacketKey(channelId, sequence)
	if err != nil {
		return false, err
	}
	valueBz := store.Get(key)
	found := len(valueBz) != 0
	return found, nil
}

// Get all pending packet sequence numbers
func (k Keeper) GetAllPendingSendPackets(ctx sdk.Context) (pendingPackets []string, err error) {
	return k.getAllPendingPackets(ctx, types.PendingSendPacketPrefix)
}

// Get all pending receive packet sequence numbers
func (k Keeper) GetAllPendingReceivePackets(ctx sdk.Context) (pendingPackets []string, err error) {
	return k.getAllPendingPackets(ctx, types.PendingReceivePacketPrefix)
}

func (k Keeper) getAllPendingPackets(ctx sdk.Context, keyPrefix []byte) (pendingPackets []string, err error) {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, keyPrefix)

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
	return k.removeAllChannelPendingPackets(ctx, types.PendingSendPacketPrefix, channelId)
}

// Remove all pending receive sequence numbers from the store
// This is executed when the quota resets
func (k Keeper) RemoveAllChannelPendingReceivePackets(ctx sdk.Context, channelId string) (err error) {
	return k.removeAllChannelPendingPackets(ctx, types.PendingReceivePacketPrefix, channelId)
}

func (k Keeper) removeAllChannelPendingPackets(ctx sdk.Context, keyPrefix []byte, channelId string) (err error) {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, keyPrefix)

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
