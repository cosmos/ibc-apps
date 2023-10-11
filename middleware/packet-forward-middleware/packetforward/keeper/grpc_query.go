package keeper

import (
	"context"

<<<<<<< HEAD:middleware/packet-forward-middleware/router/keeper/grpc_query.go
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v4/router/types"
=======
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
>>>>>>> 47f2ae0 (rename: `router` -> `packetforward` (#118)):middleware/packet-forward-middleware/packetforward/keeper/grpc_query.go

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: &params,
	}, nil
}
