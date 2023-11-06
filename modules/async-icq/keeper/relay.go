package keeper

import (
	"github.com/cosmos/ibc-apps/modules/async-icq/v7/types"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	abci "github.com/cometbft/cometbft/abci/types"

	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

// OnRecvPacket handles a given interchain queries packet on a destination host chain.
// If the transaction is successfully executed, the transaction response bytes will be returned.
func (k Keeper) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet) ([]byte, error) {
	var data types.InterchainQueryPacketData

	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		// UnmarshalJSON errors are indeterminate and therefore are not wrapped and included in failed acks
		return nil, errors.Wrapf(types.ErrUnknownDataType, "cannot unmarshal ICQ packet data")
	}

	reqs, err := types.DeserializeCosmosQuery(data.GetData())
	if err != nil {
		return nil, err
	}

	// If we panic when executing a query it should be returned as an error.
	var response []byte
	err = applyFuncIfNoError(ctx, func(ctx sdk.Context) error {
		response, err = k.executeQuery(ctx, reqs)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response, err
}

func (k Keeper) executeQuery(ctx sdk.Context, reqs []abci.RequestQuery) ([]byte, error) {
	resps := make([]abci.ResponseQuery, len(reqs))
	for i, req := range reqs {
		if err := k.authenticateQuery(ctx, req); err != nil {
			return nil, err
		}

		route := k.queryRouter.Route(req.Path)
		if route == nil {
			return nil, errors.Wrapf(sdkerrors.ErrUnauthorized, "no route found for: %s", req.Path)
		}

		resp, err := route(ctx, abci.RequestQuery{
			Data: req.Data,
			Path: req.Path,
		})
		if err != nil {
			return nil, err
		}

		// Remove non-deterministic fields from response
		resps[i] = abci.ResponseQuery{
			// Codespace is not currently part of consensus, but it will probablyy be added in the future
			// Codespace: resp.Codespace,
			Code:   resp.Code,
			Index:  resp.Index,
			Key:    resp.Key,
			Value:  resp.Value,
			Height: resp.Height,
		}
	}

	bz, err := types.SerializeCosmosResponse(resps)
	if err != nil {
		return nil, err
	}
	ack := types.InterchainQueryPacketAck{
		Data: bz,
	}
	data, err := types.ModuleCdc.MarshalJSON(&ack)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal tx data")
	}

	return data, nil
}

// authenticateQuery ensures the provided query request is in the whitelist.
func (k Keeper) authenticateQuery(ctx sdk.Context, q abci.RequestQuery) error {
	allowQueries := k.GetAllowQueries(ctx)
	if !types.ContainsQueryPath(allowQueries, q.Path) {
		return errors.Wrapf(sdkerrors.ErrUnauthorized, "query path not allowed: %s", q.Path)
	}
	if !(q.Height == 0 || q.Height == ctx.BlockHeight()) {
		return errors.Wrapf(sdkerrors.ErrUnauthorized, "query height not allowed: %d", q.Height)
	}
	if q.Prove {
		return errors.Wrapf(sdkerrors.ErrUnauthorized, "query proof not allowed")
	}

	return nil
}
