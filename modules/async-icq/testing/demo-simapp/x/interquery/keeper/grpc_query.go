package keeper

import (
	"github.com/cosmos/ibc-apps/modules/async-icq/v8/interchain-query-demo/x/interquery/types"
)

var _ types.QueryServer = Keeper{}
