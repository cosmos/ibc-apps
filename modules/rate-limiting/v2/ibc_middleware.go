package v2

import (
	"fmt"

	"github.com/cosmos/ibc-apps/modules/rate-limiting/v8/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"

	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	channeltypesv2 "github.com/cosmos/ibc-go/v10/modules/core/04-channel/v2/types"
	"github.com/cosmos/ibc-go/v10/modules/core/api"
)

var _ api.IBCModule = (*IBCMiddleware)(nil)

type IBCMiddleware struct {
	app    api.IBCModule
	keeper keeper.Keeper
}

func NewIBCMiddleware(k keeper.Keeper, app api.IBCModule) IBCMiddleware {
	return IBCMiddleware{
		app:    app,
		keeper: k,
	}
}

func (im IBCMiddleware) OnSendPacket(
	ctx sdk.Context,
	sourceClient string,
	destinationClient string,
	sequence uint64,
	payload channeltypesv2.Payload,
	signer sdk.AccAddress,
) error {
	packet := v2ToV1Packet(payload, sourceClient, destinationClient, sequence)
	if err := im.keeper.SendRateLimitedPacket(ctx, packet); err != nil {
		im.keeper.Logger(ctx).Error(fmt.Sprintf("ICS20 packet send was denied: %s", err.Error()))
		return err
	}
	return im.app.OnSendPacket(ctx, sourceClient, destinationClient, sequence, payload, signer)
}

func (im IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	sourceClient string,
	destinationClient string,
	sequence uint64,
	payload channeltypesv2.Payload,
	relayer sdk.AccAddress,
) channeltypesv2.RecvPacketResult {
	packet := v2ToV1Packet(payload, sourceClient, destinationClient, sequence)
	// Check if the packet would cause the rate limit to be exceeded,
	// and if so, return an ack error
	if err := im.keeper.ReceiveRateLimitedPacket(ctx, packet); err != nil {
		im.keeper.Logger(ctx).Error(fmt.Sprintf("ICS20 packet receive was denied: %s", err.Error()))
		return channeltypesv2.RecvPacketResult{
			Status:          channeltypesv2.PacketStatus_Failure,
			Acknowledgement: []byte(err.Error()),
		}
	}

	// If the packet was not rate-limited, pass it down to the Transfer OnRecvPacket callback
	return im.app.OnRecvPacket(ctx, sourceClient, destinationClient, sequence, payload, relayer)
}

func (im IBCMiddleware) OnTimeoutPacket(
	ctx sdk.Context,
	sourceClient string,
	destinationClient string,
	sequence uint64,
	payload channeltypesv2.Payload,
	relayer sdk.AccAddress,
) error {
	packet := v2ToV1Packet(payload, sourceClient, destinationClient, sequence)
	if err := im.keeper.TimeoutRateLimitedPacket(ctx, packet); err != nil {
		im.keeper.Logger(ctx).Error(fmt.Sprintf("ICS20 RateLimited OnTimeoutPacket failed: %s", err.Error()))
		return err
	}
	return im.app.OnTimeoutPacket(ctx, sourceClient, destinationClient, sequence, payload, relayer)
}

func (im IBCMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	sourceClient string,
	destinationClient string,
	sequence uint64,
	acknowledgement []byte,
	payload channeltypesv2.Payload,
	relayer sdk.AccAddress,
) error {
	packet := v2ToV1Packet(payload, sourceClient, destinationClient, sequence)
	if err := im.keeper.AcknowledgeRateLimitedPacket(ctx, packet, acknowledgement); err != nil {
		im.keeper.Logger(ctx).Error(fmt.Sprintf("ICS20 RateLimited OnAckPacket failed: %s", err.Error()))
		return err
	}
	return im.app.OnAcknowledgementPacket(ctx, sourceClient, destinationClient, sequence, acknowledgement, payload, relayer)
}

func v2ToV1Packet(payload channeltypesv2.Payload, sourceClient, destinationClient string, sequence uint64) channeltypes.Packet {
	return channeltypes.Packet{
		Sequence:           sequence,
		SourcePort:         payload.SourcePort,
		SourceChannel:      sourceClient,
		DestinationPort:    payload.DestinationPort,
		DestinationChannel: destinationClient,
		Data:               payload.Value,
		TimeoutHeight:      clienttypes.Height{},
		TimeoutTimestamp:   0,
	}
}
