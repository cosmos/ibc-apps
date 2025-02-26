package e2e

import (
	"context"

	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
)

func PacketAcknowledged(ctx context.Context, chain ibc.Chain, portID, channelID string, sequence uint64) bool {
	_, err := GRPCQuery[channeltypes.QueryPacketAcknowledgementResponse](ctx, chain, &channeltypes.QueryPacketAcknowledgementRequest{
		PortId:    portID,
		ChannelId: channelID,
		Sequence:  sequence,
	})
	return err == nil
}
