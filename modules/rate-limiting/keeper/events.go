package keeper

import (
	"strings"

	"github.com/cosmos/ibc-apps/modules/rate-limiting/v8/types"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// If the rate limit is exceeded or the denom is blacklisted, we emit an event
func EmitTransferDeniedEvent(ctx sdk.Context, reason, denom, channelId string, direction types.PacketDirection, amount sdkmath.Int, err error) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTransferDenied,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
			sdk.NewAttribute(types.AttributeKeyAction, strings.ToLower(direction.String())), // packet_send or packet_recv
			sdk.NewAttribute(types.AttributeKeyDenom, denom),
			sdk.NewAttribute(types.AttributeKeyChannel, channelId),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyError, err.Error()),
		),
	)
}
