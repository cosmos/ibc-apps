package keeper

import (
	"github.com/quasar-finance/interchain-query-demo/x/interquery/types"
)

var _ types.QueryServer = Keeper{}
