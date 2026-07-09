package keeper_test

import (
	"encoding/binary"

	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10/keeper"
	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"

	"cosmossdk.io/store/prefix"

	"github.com/cosmos/cosmos-sdk/runtime"
)

func (s *KeeperTestSuite) TestMigrate1to2() {
	const oldPendingSendPacketChannelLength = 16
	const oldKeyLen = oldPendingSendPacketChannelLength + 8

	writeLegacy := func(store prefix.Store, channelID string, sequence uint64) {
		key := make([]byte, oldKeyLen)
		copy(key, channelID)
		binary.BigEndian.PutUint64(key[oldPendingSendPacketChannelLength:], sequence)
		store.Set(key, []byte{1})
	}

	readAllKeys := func(store prefix.Store) [][]byte {
		it := store.Iterator(nil, nil)
		defer it.Close()

		var keys [][]byte
		for ; it.Valid(); it.Next() {
			keys = append(keys, append([]byte(nil), it.Key()...))
		}
		return keys
	}

	s.SetupTest()

	storeService := runtime.NewKVStoreService(s.App.GetKey(ratelimittypes.StoreKey))
	adapter := runtime.KVStoreAdapter(storeService.OpenKVStore(s.Ctx))
	pendingSendStore := prefix.NewStore(adapter, ratelimittypes.PendingSendPacketPrefix)
	pendingReceiveStore := prefix.NewStore(adapter, ratelimittypes.PendingReceivePacketPrefix)

	writeLegacy(pendingSendStore, "channel-1", 1)
	writeLegacy(pendingSendStore, "channel-1", 2)
	writeLegacy(pendingReceiveStore, "channel-1", 1)
	writeLegacy(pendingReceiveStore, "channel-2", 1)
	pendingSendStore.Set([]byte("unexpected-length-send"), []byte{1})
	pendingReceiveStore.Set([]byte("unexpected-length-receive"), []byte{1})

	err := s.App.RatelimitKeeper.SetPendingSendPacket(s.Ctx, "channel-3", 1, "denom-a")
	s.Require().NoError(err)
	err = s.App.RatelimitKeeper.SetPendingReceivePacket(s.Ctx, "channel-3", 1, "denom-a")
	s.Require().NoError(err)

	err = keeper.NewMigrator(s.App.RatelimitKeeper).Migrate1to2(s.Ctx)
	s.Require().NoError(err)

	s.Require().Empty(readAllKeys(pendingSendStore), "legacy pending send packet store")
	s.Require().Empty(readAllKeys(pendingReceiveStore), "legacy pending receive packet store")

	found, err := s.App.RatelimitKeeper.CheckPacketSentDuringCurrentQuota(s.Ctx, "channel-3", 1, "denom-a")
	s.Require().NoError(err)
	s.Require().True(found, "collections pending send packet should be preserved")
	found, err = s.App.RatelimitKeeper.CheckPacketReceivedDuringCurrentQuota(s.Ctx, "channel-3", 1, "denom-a")
	s.Require().NoError(err)
	s.Require().True(found, "collections pending receive packet should be preserved")
}
