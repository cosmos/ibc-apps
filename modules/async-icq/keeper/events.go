package keeper

import (
	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v7/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ibc-go/v7/modules/core/exported"
)

// EmitWriteErrorAcknowledgementEvent emits an event signalling an error acknowledgement and including the error details
func EmitWriteErrorAcknowledgementEvent(ctx sdk.Context, packet exported.PacketI, err error) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			icqtypes.EventTypePacketError,
			sdk.NewAttribute(sdk.AttributeKeyModule, icqtypes.ModuleName),
			sdk.NewAttribute(icqtypes.AttributeKeyAckError, err.Error()),
			sdk.NewAttribute(icqtypes.AttributeKeyHostChannelID, packet.GetDestChannel()),
		),
	)
}
