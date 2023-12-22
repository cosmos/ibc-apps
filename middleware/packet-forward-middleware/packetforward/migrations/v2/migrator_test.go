package v2_test

import (
	"testing"

	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/exported"
	v2 "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/migrations/v2"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

// 0.02
var migrateExpected = sdk.NewDecWithPrec(2, 2)

type mockSubspace struct {
	ps types.Params
}

func newMockSubspace(ps types.Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSet(_ sdk.Context, ps exported.ParamSet) {
	*ps.(*types.Params) = ms.ps
}

// TestMigrate validates the legacySubstore value moves to the new params key.
func TestMigrate(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(packetforward.AppModuleBasic{})
	cdc := encCfg.Codec

	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)

	legacySubspace := newMockSubspace(types.Params{
		FeePercentage: migrateExpected,
	})
	require.NoError(t, v2.Migrate(ctx, store, legacySubspace, cdc))

	var res types.Params
	bz := store.Get(types.ParamsKey)
	require.NoError(t, cdc.Unmarshal(bz, &res))
	require.Equal(t, legacySubspace.ps, res)
}
