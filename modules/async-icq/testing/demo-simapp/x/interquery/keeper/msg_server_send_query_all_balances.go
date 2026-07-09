package keeper

import (
	"context"
	"time"

	"github.com/cosmos/ibc-apps/modules/async-icq/v8/interchain-query-demo/x/interquery/types"
	clienttypes "github.com/cosmos/ibc-go/v11/modules/core/02-client/types"
	abcitypes "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (k msgServer) SendQueryAllBalances(goCtx context.Context, msg *types.MsgSendQueryAllBalances) (*types.MsgSendQueryAllBalancesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	q := banktypes.QueryAllBalancesRequest{
		Address:    msg.Address,
		Pagination: msg.Pagination,
	}
	reqs := []abcitypes.RequestQuery{
		{
			Path: "/cosmos.bank.v1beta1.Query/AllBalances",
			Data: k.cdc.MustMarshal(&q),
		},
	}

	// timeoutTimestamp set to max value with the unsigned bit shifted to satisfy hermes timestamp conversion
	// it is the responsibility of the auth module developer to ensure an appropriate timeout timestamp
	timeoutTimestamp := ctx.BlockTime().Add(time.Minute).UnixNano()
	seq, err := k.SendQuery(ctx, types.PortID, msg.ChannelId, reqs, clienttypes.ZeroHeight(), uint64(timeoutTimestamp))
	if err != nil {
		return nil, err
	}

	k.SetQueryRequest(ctx, seq, q)

	return &types.MsgSendQueryAllBalancesResponse{
		Sequence: seq,
	}, nil
}
