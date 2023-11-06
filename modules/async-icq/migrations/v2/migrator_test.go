package v2_test

import (
	"testing"

	icq "github.com/cosmos/ibc-apps/modules/async-icq/v7"
	"github.com/cosmos/ibc-apps/modules/async-icq/v7/exported"
	v2 "github.com/cosmos/ibc-apps/modules/async-icq/v7/migrations/v2"
	"github.com/cosmos/ibc-apps/modules/async-icq/v7/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

var (
	// it is on by default, so we check the false case
	hostEnabled    = false
	allowedQueries = []string{"/cosmos.bank.v1beta1.Query/AllBalances"}
)

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
	encCfg := moduletestutil.MakeTestEncodingConfig(icq.AppModuleBasic{})
	cdc := encCfg.Codec

	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)

	legacySubspace := newMockSubspace(types.Params{
		HostEnabled:  hostEnabled,
		AllowQueries: allowedQueries,
	})
	require.NoError(t, v2.Migrate(ctx, store, legacySubspace, cdc))

	var res types.Params
	bz := store.Get(types.ParamsKey)
	require.NoError(t, cdc.Unmarshal(bz, &res))
	require.Equal(t, legacySubspace.ps, res)
}
