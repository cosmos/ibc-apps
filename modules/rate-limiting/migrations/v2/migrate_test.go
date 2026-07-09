package v2_test

import (
	"testing"

	v2 "github.com/cosmos/ibc-apps/modules/rate-limiting/v10/migrations/v2"
	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10/testing/simapp/apptesting"
	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/store/prefix"

	"github.com/cosmos/cosmos-sdk/runtime"
)

func TestMigrateClearsLegacyPendingPacketStores(t *testing.T) {
	helper := apptesting.AppTestHelper{}
	helper.Setup()
	ctx := helper.Ctx

	storeService := runtime.NewKVStoreService(helper.App.GetKey(types.StoreKey))
	adapter := runtime.KVStoreAdapter(storeService.OpenKVStore(ctx))
	pendingSendStore := prefix.NewStore(adapter, types.PendingSendPacketPrefix)
	pendingReceiveStore := prefix.NewStore(adapter, types.PendingReceivePacketPrefix)

	pendingSendStore.Set([]byte("send-1"), []byte{1})
	pendingSendStore.Set([]byte("send-2"), []byte{2})
	pendingReceiveStore.Set([]byte("receive-1"), []byte{1})
	pendingReceiveStore.Set([]byte("receive-2"), []byte{2})

	require.NoError(t, v2.Migrate(ctx, storeService))
	require.Empty(t, collectKeys(t, pendingSendStore))
	require.Empty(t, collectKeys(t, pendingReceiveStore))
}

func collectKeys(t *testing.T, store prefix.Store) [][]byte {
	t.Helper()

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var keys [][]byte
	for ; iterator.Valid(); iterator.Next() {
		keys = append(keys, append([]byte(nil), iterator.Key()...))
	}

	return keys
}
