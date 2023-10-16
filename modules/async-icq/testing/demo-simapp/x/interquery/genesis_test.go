package interquery_test

import (
	"testing"

	keepertest "github.com/quasar-finance/interchain-query-demo/testutil/keeper"
	"github.com/quasar-finance/interchain-query-demo/testutil/nullify"
	"github.com/quasar-finance/interchain-query-demo/x/interquery"
	"github.com/quasar-finance/interchain-query-demo/x/interquery/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		PortId: types.PortID,
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.InterqueryKeeper(t)
	interquery.InitGenesis(ctx, *k, genesisState)
	got := interquery.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.Equal(t, genesisState.PortId, got.PortId)

	// this line is used by starport scaffolding # genesis/test/assert
}
