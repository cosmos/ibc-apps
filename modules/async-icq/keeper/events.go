package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ibc-go/v5/modules/core/exported"
	icqtypes "github.com/strangelove-ventures/async-icq/v5/types"
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
