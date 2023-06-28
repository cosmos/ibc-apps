package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"cosmossdk.io/math"
	"github.com/cosmos/ibc-apps/middleware/staking-middleware/v7/staking/keeper"
	"github.com/cosmos/ibc-apps/middleware/staking-middleware/v7/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"

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
}

// NewIBCMiddleware creates a new IBCMiddleware given the keeper and underlying application.
func NewIBCMiddleware(
	app porttypes.IBCModule,
	k *keeper.Keeper,
) IBCMiddleware {
	return IBCMiddleware{
		app:    app,
		keeper: k,
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

// NOTE: prob still need this
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
		return channeltypes.NewErrorAcknowledgement(err)
	}

	_, sbz, err := bech32.DecodeAndConvert(data.Sender)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	_, rbz, err := bech32.DecodeAndConvert(data.Receiver)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	if !bytes.Equal(sbz, rbz) {
		return channeltypes.NewErrorAcknowledgement(fmt.Errorf("staking middleware only staking from the same address as was sent"))
	}

	im.keeper.Logger(ctx).Debug("stakingMiddleWare OnRecvPacket",
		"sequence", packet.Sequence,
		"src-channel", packet.SourceChannel, "src-port", packet.SourcePort,
		"dst-channel", packet.DestinationChannel, "dst-port", packet.DestinationPort,
		"amount", data.Amount, "denom", data.Denom, "memo", data.Memo,
	)

	d := make(map[string]interface{})

	if err = json.Unmarshal([]byte(data.Memo), &d); err != nil || d["stake"] == nil {
		// not a packet that should be staked
		im.keeper.Logger(ctx).Debug("stakingMiddleware OnRecvPacket staking metadata does not exist")
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}
	m := &types.PacketMetadata{}
	err = json.Unmarshal([]byte(data.Memo), m)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Errorf("stakingMiddleware error parsing staking metadata, %s", err))
	}

	// validate middleware args
	if err := m.Stake.Validate(); err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// if the denom sent isn't the bond denom for the chain then do this
	bondDenom := im.keeper.StakingKeeper.BondDenom(ctx)
	if getDenomForThisChain(packet.DestinationPort, packet.DestinationChannel, packet.SourcePort, packet.SourceChannel, data.Denom) != bondDenom {
		return channeltypes.NewErrorAcknowledgement(fmt.Errorf("invalid coin denomination: got %s, expected %s", data.Denom, bondDenom))
	}

	// ensure that stake amount isn't greater than amount sent
	packetAmount, _ := math.NewIntFromString(data.Amount)
	stakeAmount := m.Stake.AmountInt()
	if stakeAmount.GT(packetAmount) {
		return channeltypes.NewErrorAcknowledgement(fmt.Errorf("stake amount (%s) greater than amount sent over (%s)", m.Stake.StakeAmount, data.Amount))
	}

	// special case 0 amount to mean whole packet
	var delegationAmount math.Int
	if stakeAmount.IsZero() {
		delegationAmount = packetAmount
	} else {
		delegationAmount = stakeAmount
	}
	// make sure that the validator exists on this chain
	validator, found := im.keeper.StakingKeeper.GetValidator(ctx, m.Stake.ValAddr())
	if !found {
		return channeltypes.NewErrorAcknowledgement(stakingTypes.ErrNoValidatorFound)
	}

	delegatorAddress, err := sdk.AccAddressFromBech32(data.Receiver)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Errorf("stakingMiddleware invalid address, %s", err))
	}

	ack := im.app.OnRecvPacket(ctx, packet, relayer)
	if !ack.Success() {
		return ack
	}

	newShares, err := im.keeper.StakingKeeper.Delegate(ctx, delegatorAddress, delegationAmount, stakingTypes.Unbonded, validator, true)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Errorf("error bonding tokens: %w", err))
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			stakingTypes.EventTypeDelegate,
			sdk.NewAttribute(stakingTypes.AttributeKeyValidator, m.Stake.ValidatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, m.Stake.StakeAmount),
			sdk.NewAttribute(stakingTypes.AttributeKeyNewShares, newShares.String()),
		),
	})

	// returning nil ack will prevent WriteAcknowledgement from occurring for forwarded packet.
	// This is intentional so that the acknowledgement will be written later based on the ack/timeout of the forwarded packet.
	return ack
}

// OnAcknowledgementPacket implements the IBCModule interface.
func (im IBCMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	// I don't think we need to do anything here.
	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements the IBCModule interface.
func (im IBCMiddleware) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	// I don't think we need to do anything here.
	return im.app.OnTimeoutPacket(ctx, packet, relayer)
}

func (im IBCMiddleware) GetAppVersion(
	ctx sdk.Context,
	portID,
	channelID string,
) (string, bool) {
	return im.keeper.ICS4Wrapper.GetAppVersion(ctx, portID, channelID)
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
	return im.keeper.ICS4Wrapper.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
}

// WriteAcknowledgement implements the ICS4 Wrapper interface.
func (im IBCMiddleware) WriteAcknowledgement(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet ibcexported.PacketI,
	ack ibcexported.Acknowledgement,
) error {
	return im.keeper.ICS4Wrapper.WriteAcknowledgement(ctx, chanCap, packet, ack)
}
