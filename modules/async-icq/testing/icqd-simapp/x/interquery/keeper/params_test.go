package keeper_test

import (
	"testing"

	testkeeper "github.com/quasar-finance/interchain-query-demo/testutil/keeper"
	"github.com/quasar-finance/interchain-query-demo/x/interquery/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.InterqueryKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
