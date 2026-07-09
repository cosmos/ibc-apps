package keeper

import (
	"github.com/cosmos/ibc-apps/modules/async-icq/v8/interchain-query-demo/x/interquery/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	abci "github.com/cometbft/cometbft/abci/types"
	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v8/types"
	clienttypes "github.com/cosmos/ibc-go/v11/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v11/modules/core/04-channel/types"
)

// SendQuery serializes the given ABCI query requests into an interchain-query
// packet and sends it over the given channel using the ICS4Wrapper (ibc-go v11).
func (k Keeper) SendQuery(
	ctx sdk.Context,
	sourcePort,
	sourceChannel string,
	reqs []abci.RequestQuery,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) (uint64, error) {
	data, err := icqtypes.SerializeCosmosQuery(reqs)
	if err != nil {
		return 0, errorsmod.Wrap(err, "could not serialize reqs into cosmos query")
	}
	icqPacketData := icqtypes.InterchainQueryPacketData{
		Data: data,
	}
	if err := icqPacketData.ValidateBasic(); err != nil {
		return 0, errorsmod.Wrap(err, "invalid interchain query packet data")
	}

	return k.ics4Wrapper.SendPacket(
		ctx,
		sourcePort,
		sourceChannel,
		timeoutHeight,
		timeoutTimestamp,
		icqPacketData.GetBytes(),
	)
}

func (k Keeper) OnAcknowledgementPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	ack channeltypes.Acknowledgement,
) error {
	switch resp := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Result:
		var ackData icqtypes.InterchainQueryPacketAck
		if err := icqtypes.ModuleCdc.UnmarshalJSON(resp.Result, &ackData); err != nil {
			return errorsmod.Wrap(err, "failed to unmarshal interchain query packet ack")
		}
		resps, err := icqtypes.DeserializeCosmosResponse(ackData.Data)
		if err != nil {
			return errorsmod.Wrap(err, "could not deserialize data to cosmos response")
		}

		if len(resps) < 1 {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "no responses in interchain query packet ack")
		}

		var r banktypes.QueryAllBalancesResponse
		if err := k.cdc.Unmarshal(resps[0].Value, &r); err != nil {
			return errorsmod.Wrapf(err, "failed to unmarshal interchain query response to type %T", resp)
		}

		k.SetQueryResponse(ctx, modulePacket.Sequence, r)
		k.SetLastQueryPacketSeq(ctx, modulePacket.Sequence)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeQueryResult,
				sdk.NewAttribute(types.AttributeKeyAckSuccess, string(resp.Result)),
			),
		)

		k.Logger(ctx).Info("interchain query response", "sequence", modulePacket.Sequence, "response", r)
	case *channeltypes.Acknowledgement_Error:
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeQueryResult,
				sdk.NewAttribute(types.AttributeKeyAckError, resp.Error),
			),
		)

		k.Logger(ctx).Error("interchain query response", "sequence", modulePacket.Sequence, "error", resp.Error)
	}
	return nil
}
