package dummyware

import (
	"context"

	forwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v9/packetforward/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	channeltypes "github.com/cosmos/ibc-go/v9/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v9/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v9/modules/core/exported"
)

var _ porttypes.IBCModule = &IBCMiddleware{}

// IBCMiddleware implements the ICS26 callbacks for the dummy middleware given the
// dummy keeper and the underlying application.
type IBCMiddleware struct {
	app porttypes.IBCModule
}

// NewIBCMiddleware creates a new IBCMiddleware given the keeper and underlying application.
func NewIBCMiddleware(app porttypes.IBCModule) IBCMiddleware {
	return IBCMiddleware{
		app: app,
	}
}

// OnChanOpenInit implements the IBCModule interface.
func (im IBCMiddleware) OnChanOpenInit(
	ctx context.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	return im.app.OnChanOpenInit(ctx, order, connectionHops, portID, channelID, counterparty, version)
}

// OnChanOpenTry implements the IBCModule interface.
func (im IBCMiddleware) OnChanOpenTry(
	ctx context.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID, channelID string,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (version string, err error) {
	return im.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, counterparty, counterpartyVersion)
}

// OnChanOpenAck implements the IBCModule interface.
func (im IBCMiddleware) OnChanOpenAck(
	ctx context.Context,
	portID, channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

// OnChanOpenConfirm implements the IBCModule interface.
func (im IBCMiddleware) OnChanOpenConfirm(ctx context.Context, portID, channelID string) error {
	return im.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanCloseInit implements the IBCModule interface.
func (im IBCMiddleware) OnChanCloseInit(ctx context.Context, portID, channelID string) error {
	return im.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanCloseConfirm implements the IBCModule interface.
func (im IBCMiddleware) OnChanCloseConfirm(ctx context.Context, portID, channelID string) error {
	return im.app.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnRecvPacket sets the non-refundable context value to true so that on a failed forward funds will not
// be refunded from the intermediate chain.
func (im IBCMiddleware) OnRecvPacket(
	ctx context.Context,
	channelVersion string,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	// Compose our context with values that will be used to pass through to the forward middleware
	ctxWithFlags := context.WithValue(ctx, forwardtypes.NonrefundableKey{}, true)

	// Call into underlying app to receive funds on this chain
	return im.app.OnRecvPacket(ctxWithFlags, channelVersion, packet, relayer)
}

// OnAcknowledgementPacket implements the IBCModule interface.
func (im IBCMiddleware) OnAcknowledgementPacket(
	ctx context.Context,
	channelVersion string,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	return im.app.OnAcknowledgementPacket(ctx, channelVersion, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements the IBCModule interface.
func (im IBCMiddleware) OnTimeoutPacket(ctx context.Context, channelVersion string, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	return im.app.OnTimeoutPacket(ctx, channelVersion, packet, relayer)
}
