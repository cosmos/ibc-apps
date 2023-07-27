package router

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/armon/go-metrics"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/router/keeper"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/router/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
)

var _ porttypes.Middleware = &IBCMiddleware{}

// IBCMiddleware implements the ICS26 callbacks for the forward middleware given the
// forward keeper and the underlying application.
type IBCMiddleware struct {
	app    porttypes.IBCModule
	keeper *keeper.Keeper

	retriesOnTimeout uint8
	forwardTimeout   time.Duration
	refundTimeout    time.Duration
}

// NewIBCMiddleware creates a new IBCMiddleware given the keeper and underlying application.
func NewIBCMiddleware(
	app porttypes.IBCModule,
	k *keeper.Keeper,
	retriesOnTimeout uint8,
	forwardTimeout time.Duration,
	refundTimeout time.Duration,
) IBCMiddleware {
	return IBCMiddleware{
		app:              app,
		keeper:           k,
		retriesOnTimeout: retriesOnTimeout,
		forwardTimeout:   forwardTimeout,
		refundTimeout:    refundTimeout,
	}
}

// OnChanOpenInit implements the IBCModule interface.
func (im IBCMiddleware) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	return im.app.OnChanOpenInit(ctx, order, connectionHops, portID, channelID, chanCap, counterparty, version)
}

// OnChanOpenTry implements the IBCModule interface.
func (im IBCMiddleware) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID, channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (version string, err error) {
	return im.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, chanCap, counterparty, counterpartyVersion)
}

// OnChanOpenAck implements the IBCModule interface.
func (im IBCMiddleware) OnChanOpenAck(
	ctx sdk.Context,
	portID, channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

// OnChanOpenConfirm implements the IBCModule interface.
func (im IBCMiddleware) OnChanOpenConfirm(ctx sdk.Context, portID, channelID string) error {
	return im.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanCloseInit implements the IBCModule interface.
func (im IBCMiddleware) OnChanCloseInit(ctx sdk.Context, portID, channelID string) error {
	return im.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanCloseConfirm implements the IBCModule interface.
func (im IBCMiddleware) OnChanCloseConfirm(ctx sdk.Context, portID, channelID string) error {
	return im.app.OnChanCloseConfirm(ctx, portID, channelID)
}

func getDenomForThisChain(port, channel, counterpartyPort, counterpartyChannel, denom string) string {
	counterpartyPrefix := transfertypes.GetDenomPrefix(counterpartyPort, counterpartyChannel)
	if strings.HasPrefix(denom, counterpartyPrefix) {
		// unwind denom
		unwoundDenom := denom[len(counterpartyPrefix):]
		denomTrace := transfertypes.ParseDenomTrace(unwoundDenom)
		if denomTrace.Path == "" {
			// denom is now unwound back to native denom
			return unwoundDenom
		}
		// denom is still IBC denom
		return denomTrace.IBCDenom()
	}
	// append port and channel from this chain to denom
	prefixedDenom := transfertypes.GetDenomPrefix(port, channel) + denom
	return transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()
}

// getBoolFromAny returns the bool value is any is a valid bool, otherwise false.
func getBoolFromAny(value any) bool {
	if value == nil {
		return false
	}
	boolVal, ok := value.(bool)
	if !ok {
		return false
	}
	return boolVal
}

// getReceiver returns the receiver address for a given channel and original sender.
// it overrides the receiver address to be a hash of the channel/origSender so that
// the receiver address is deterministic and can be used to identify the sender on the
// initial chain.
func getReceiver(channel string, originalSender string) (string, error) {
	senderStr := fmt.Sprintf("%s/%s", channel, originalSender)
	senderHash32 := address.Hash(types.ModuleName, []byte(senderStr))
	sender := sdk.AccAddress(senderHash32[:20])
	return sdk.Bech32ifyAddressBytes(sdk.Bech32MainPrefix, sender)
}

// IMPORTANT: ensure errors are deterministic, otherwise the app hash will be non-deterministic across validators.
func newErrorAcknowledgement(err error) channeltypes.Acknowledgement {
	return channeltypes.Acknowledgement{
		Response: &channeltypes.Acknowledgement_Error{
			Error: fmt.Sprintf("packet-forward-middleware error: %s", err.Error()),
		},
	}
}

// OnRecvPacket checks the memo field on this packet and if the metadata inside's root key indicates this packet
// should be handled by the swap middleware it attempts to perform a swap. If the swap is successful
// the underlying application's OnRecvPacket callback is invoked, an ack error is returned otherwise.
func (im IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return newErrorAcknowledgement(fmt.Errorf("failed to unmarshal packet data as FungibleTokenPacketData: %s", err.Error()))
	}

	im.keeper.Logger(ctx).Debug("packetForwardMiddleware OnRecvPacket",
		"sequence", packet.Sequence,
		"src-channel", packet.SourceChannel, "src-port", packet.SourcePort,
		"dst-channel", packet.DestinationChannel, "dst-port", packet.DestinationPort,
		"amount", data.Amount, "denom", data.Denom, "memo", data.Memo,
	)

	d := make(map[string]interface{})
	err := json.Unmarshal([]byte(data.Memo), &d)
	if err != nil || d["forward"] == nil {
		// not a packet that should be forwarded
		im.keeper.Logger(ctx).Debug("packetForwardMiddleware OnRecvPacket forward metadata does not exist")
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}
	m := &types.PacketMetadata{}
	err = json.Unmarshal([]byte(data.Memo), m)
	if err != nil {
		return newErrorAcknowledgement(fmt.Errorf("error parsing forward metadata: %s", err.Error()))
	}

	metadata := m.Forward

	goCtx := ctx.Context()
	processed := getBoolFromAny(goCtx.Value(types.ProcessedKey{}))
	nonrefundable := getBoolFromAny(goCtx.Value(types.NonrefundableKey{}))
	disableDenomComposition := getBoolFromAny(goCtx.Value(types.DisableDenomCompositionKey{}))

	if err := metadata.Validate(); err != nil {
		return newErrorAcknowledgement(err)
	}

	// override the receiver so that senders cannot move funds through arbitrary addresses.
	overrideReceiver, err := getReceiver(packet.DestinationChannel, data.Sender)
	if err != nil {
		return newErrorAcknowledgement(fmt.Errorf("failed to construct override receiver: %s", err.Error()))
	}

	// if this packet has been handled by another middleware in the stack there may be no need to call into the
	// underlying app, otherwise the transfer module's OnRecvPacket callback could be invoked more than once
	// which would mint/burn vouchers more than once
	if !processed {
		data.Receiver = overrideReceiver
		packet.Data = transfertypes.ModuleCdc.MustMarshalJSON(&data)
		ack := im.app.OnRecvPacket(ctx, packet, relayer)
		if ack == nil || !ack.Success() {
			return ack
		}
	}

	// if this packet's token denom is already the base denom for some native token on this chain,
	// we do not need to do any further composition of the denom before forwarding the packet
	denomOnThisChain := data.Denom
	if !disableDenomComposition {
		denomOnThisChain = getDenomForThisChain(
			packet.DestinationPort, packet.DestinationChannel,
			packet.SourcePort, packet.SourceChannel,
			data.Denom,
		)
	}

	amountInt, ok := sdk.NewIntFromString(data.Amount)
	if !ok {
		return newErrorAcknowledgement(fmt.Errorf("error parsing amount for forward: %s", data.Amount))
	}

	token := sdk.NewCoin(denomOnThisChain, amountInt)

	timeout := time.Duration(metadata.Timeout)

	if timeout.Nanoseconds() <= 0 {
		timeout = im.forwardTimeout
	}

	var retries uint8
	if metadata.Retries != nil {
		retries = *metadata.Retries
	} else {
		retries = im.retriesOnTimeout
	}

	err = im.keeper.ForwardTransferPacket(ctx, nil, packet, data.Sender, overrideReceiver, metadata, token, retries, timeout, []metrics.Label{}, nonrefundable)
	if err != nil {
		return newErrorAcknowledgement(err)
	}

	// returning nil ack will prevent WriteAcknowledgement from occurring for forwarded packet.
	// This is intentional so that the acknowledgement will be written later based on the ack/timeout of the forwarded packet.
	return nil
}

// OnAcknowledgementPacket implements the IBCModule interface.
func (im IBCMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		im.keeper.Logger(ctx).Error("packetForwardMiddleware error parsing packet data from ack packet",
			"sequence", packet.Sequence,
			"src-channel", packet.SourceChannel, "src-port", packet.SourcePort,
			"dst-channel", packet.DestinationChannel, "dst-port", packet.DestinationPort,
			"error", err,
		)
		return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	im.keeper.Logger(ctx).Debug("packetForwardMiddleware OnAcknowledgementPacket",
		"sequence", packet.Sequence,
		"src-channel", packet.SourceChannel, "src-port", packet.SourcePort,
		"dst-channel", packet.DestinationChannel, "dst-port", packet.DestinationPort,
		"amount", data.Amount, "denom", data.Denom,
	)

	var ack channeltypes.Acknowledgement
	if err := channeltypes.SubModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}

	inFlightPacket := im.keeper.GetAndClearInFlightPacket(ctx, packet.SourceChannel, packet.SourcePort, packet.Sequence)
	if inFlightPacket != nil {
		// this is a forwarded packet, so override handling to avoid refund from being processed.
		return im.keeper.WriteAcknowledgementForForwardedPacket(ctx, packet, data, inFlightPacket, ack)
	}

	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements the IBCModule interface.
func (im IBCMiddleware) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		im.keeper.Logger(ctx).Error("packetForwardMiddleware error parsing packet data from timeout packet",
			"sequence", packet.Sequence,
			"src-channel", packet.SourceChannel, "src-port", packet.SourcePort,
			"dst-channel", packet.DestinationChannel, "dst-port", packet.DestinationPort,
			"error", err,
		)
		return im.app.OnTimeoutPacket(ctx, packet, relayer)
	}

	im.keeper.Logger(ctx).Debug("packetForwardMiddleware OnAcknowledgementPacket",
		"sequence", packet.Sequence,
		"src-channel", packet.SourceChannel, "src-port", packet.SourcePort,
		"dst-channel", packet.DestinationChannel, "dst-port", packet.DestinationPort,
		"amount", data.Amount, "denom", data.Denom,
	)

	inFlightPacket, err := im.keeper.TimeoutShouldRetry(ctx, packet)
	if inFlightPacket != nil {
		if err != nil {
			im.keeper.RemoveInFlightPacket(ctx, packet)
			// this is a forwarded packet, so override handling to avoid refund from being processed on this chain.
			// WriteAcknowledgement with proxied ack to return success/fail to previous chain.
			return im.keeper.WriteAcknowledgementForForwardedPacket(ctx, packet, data, inFlightPacket, newErrorAcknowledgement(err))
		}
		// timeout should be retried. In order to do that, we need to handle this timeout to refund on this chain first.
		if err := im.app.OnTimeoutPacket(ctx, packet, relayer); err != nil {
			return err
		}
		return im.keeper.RetryTimeout(ctx, packet.SourceChannel, packet.SourcePort, data, inFlightPacket)
	}

	return im.app.OnTimeoutPacket(ctx, packet, relayer)
}

// SendPacket implements the ICS4 Wrapper interface.
func (im IBCMiddleware) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string, sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	return im.keeper.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
}

// WriteAcknowledgement implements the ICS4 Wrapper interface.
func (im IBCMiddleware) WriteAcknowledgement(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet ibcexported.PacketI,
	ack ibcexported.Acknowledgement,
) error {
	return im.keeper.WriteAcknowledgement(ctx, chanCap, packet, ack)
}

func (im IBCMiddleware) GetAppVersion(
	ctx sdk.Context,
	portID,
	channelID string,
) (string, bool) {
	return im.keeper.GetAppVersion(ctx, portID, channelID)
}
