package keeper

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/armon/go-metrics"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/router/types"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/cometbft/cometbft/libs/log"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	coretypes "github.com/cosmos/ibc-go/v7/modules/core/types"
)

var (
	// DefaultTransferPacketTimeoutHeight is the timeout height following IBC defaults
	DefaultTransferPacketTimeoutHeight = clienttypes.Height{
		RevisionNumber: 0,
		RevisionHeight: 0,
	}

	// DefaultForwardTransferPacketTimeoutTimestamp is the timeout timestamp following IBC defaults
	DefaultForwardTransferPacketTimeoutTimestamp = time.Duration(transfertypes.DefaultRelativePacketTimeoutTimestamp) * time.Nanosecond

	// DefaultRefundTransferPacketTimeoutTimestamp is a 28-day timeout for refund packets since funds are stuck in router module otherwise.
	DefaultRefundTransferPacketTimeoutTimestamp = 28 * 24 * time.Hour
)

// Keeper defines the packet forward middleware keeper
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	paramSpace paramtypes.Subspace

	transferKeeper types.TransferKeeper
	channelKeeper  types.ChannelKeeper
	distrKeeper    types.DistributionKeeper
	bankKeeper     types.BankKeeper
	ics4Wrapper    porttypes.ICS4Wrapper
}

// NewKeeper creates a new forward Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	paramSpace paramtypes.Subspace,
	transferKeeper types.TransferKeeper,
	channelKeeper types.ChannelKeeper,
	distrKeeper types.DistributionKeeper,
	bankKeeper types.BankKeeper,
	ics4Wrapper porttypes.ICS4Wrapper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:            cdc,
		storeKey:       key,
		transferKeeper: transferKeeper,
		channelKeeper:  channelKeeper,
		paramSpace:     paramSpace,
		distrKeeper:    distrKeeper,
		bankKeeper:     bankKeeper,
		ics4Wrapper:    ics4Wrapper,
	}
}

// SetTransferKeeper sets the transferKeeper
func (k *Keeper) SetTransferKeeper(transferKeeper types.TransferKeeper) {
	k.transferKeeper = transferKeeper
}

// Logger returns a module-specific logger.
func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+ibcexported.ModuleName+"-"+types.ModuleName)
}

func (k *Keeper) WriteAcknowledgementForForwardedPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data transfertypes.FungibleTokenPacketData,
	inFlightPacket *types.InFlightPacket,
	ack channeltypes.Acknowledgement,
) error {
	// Lookup module by channel capability
	_, chanCap, err := k.channelKeeper.LookupModuleByChannel(ctx, inFlightPacket.RefundPortId, inFlightPacket.RefundChannelId)
	if err != nil {
		return errorsmod.Wrap(err, "could not retrieve module from port-id")
	}

	// for forwarded packets, the funds were moved into an escrow account if the denom originated on this chain.
	// On an ack error or timeout on a forwarded packet, the funds in the escrow account
	// should be moved to the other escrow account on the other side or burned.
	if !ack.Success() {
		// If this packet is non-refundable due to some action that took place between the initial ibc transfer and the forward
		// we write a successful ack containing details on what happened regardless of ack error or timeout
		if inFlightPacket.Nonrefundable {
			ackResult := fmt.Sprintf("packet forward failed after point of no return: %s", ack.GetError())
			newAck := channeltypes.NewResultAcknowledgement([]byte(ackResult))

			return k.ics4Wrapper.WriteAcknowledgement(ctx, chanCap, channeltypes.Packet{
				Data:               inFlightPacket.PacketData,
				Sequence:           inFlightPacket.RefundSequence,
				SourcePort:         inFlightPacket.PacketSrcPortId,
				SourceChannel:      inFlightPacket.PacketSrcChannelId,
				DestinationPort:    inFlightPacket.RefundPortId,
				DestinationChannel: inFlightPacket.RefundChannelId,
				TimeoutHeight:      clienttypes.MustParseHeight(inFlightPacket.PacketTimeoutHeight),
				TimeoutTimestamp:   inFlightPacket.PacketTimeoutTimestamp,
			}, newAck)
		}

		fullDenomPath := data.Denom

		// deconstruct the token denomination into the denomination trace info
		// to determine if the sender is the source chain
		if strings.HasPrefix(data.Denom, "ibc/") {
			fullDenomPath, err = k.transferKeeper.DenomPathFromHash(ctx, data.Denom)
			if err != nil {
				return err
			}
		}

		if transfertypes.SenderChainIsSource(packet.SourcePort, packet.SourceChannel, fullDenomPath) {
			// funds were moved to escrow account for transfer, so they need to either:
			// - move to the other escrow account, in the case of native denom
			// - burn

			amount, ok := sdk.NewIntFromString(data.Amount)
			if !ok {
				return fmt.Errorf("failed to parse amount from packet data for forward refund: %s", data.Amount)
			}
			denomTrace := transfertypes.ParseDenomTrace(fullDenomPath)
			token := sdk.NewCoin(denomTrace.IBCDenom(), amount)

			escrowAddress := transfertypes.GetEscrowAddress(packet.SourcePort, packet.SourceChannel)

			if transfertypes.SenderChainIsSource(inFlightPacket.RefundPortId, inFlightPacket.RefundChannelId, fullDenomPath) {
				// transfer funds from escrow account for forwarded packet to escrow account going back for refund.

				refundEscrowAddress := transfertypes.GetEscrowAddress(inFlightPacket.RefundPortId, inFlightPacket.RefundChannelId)

				if err := k.bankKeeper.SendCoins(
					ctx, escrowAddress, refundEscrowAddress, sdk.NewCoins(token),
				); err != nil {
					return fmt.Errorf("failed to send coins from escrow account to refund escrow account: %w", err)
				}
			} else {
				// transfer the coins from the escrow account to the module account and burn them.

				if err := k.bankKeeper.SendCoinsFromAccountToModule(
					ctx, escrowAddress, transfertypes.ModuleName, sdk.NewCoins(token),
				); err != nil {
					return fmt.Errorf("failed to send coins from escrow to module account for burn: %w", err)
				}

				if err := k.bankKeeper.BurnCoins(
					ctx, transfertypes.ModuleName, sdk.NewCoins(token),
				); err != nil {
					// NOTE: should not happen as the module account was
					// retrieved on the step above and it has enough balace
					// to burn.
					panic(fmt.Sprintf("cannot burn coins after a successful send from escrow account to module account: %v", err))
				}
			}

			// We move funds from the escrowAddress in both cases,
			// update the total escrow amount for the denom.
			k.unescrowToken(ctx, token)
		}
	}

	return k.ics4Wrapper.WriteAcknowledgement(ctx, chanCap, channeltypes.Packet{
		Data:               inFlightPacket.PacketData,
		Sequence:           inFlightPacket.RefundSequence,
		SourcePort:         inFlightPacket.PacketSrcPortId,
		SourceChannel:      inFlightPacket.PacketSrcChannelId,
		DestinationPort:    inFlightPacket.RefundPortId,
		DestinationChannel: inFlightPacket.RefundChannelId,
		TimeoutHeight:      clienttypes.MustParseHeight(inFlightPacket.PacketTimeoutHeight),
		TimeoutTimestamp:   inFlightPacket.PacketTimeoutTimestamp,
	}, ack)
}

// unescrowToken will update the total escrow by deducting the unescrowed token
// from the current total escrow.
func (k *Keeper) unescrowToken(ctx sdk.Context, token sdk.Coin) {
	currentTotalEscrow := k.transferKeeper.GetTotalEscrowForDenom(ctx, token.GetDenom())
	newTotalEscrow := currentTotalEscrow.Sub(token)
	k.transferKeeper.SetTotalEscrowForDenom(ctx, newTotalEscrow)
}

func (k *Keeper) ForwardTransferPacket(
	ctx sdk.Context,
	inFlightPacket *types.InFlightPacket,
	srcPacket channeltypes.Packet,
	srcPacketSender string,
	receiver string,
	metadata *types.ForwardMetadata,
	token sdk.Coin,
	maxRetries uint8,
	timeout time.Duration,
	labels []metrics.Label,
	nonrefundable bool,
) error {
	var err error
	feeAmount := sdk.NewDecFromInt(token.Amount).Mul(k.GetFeePercentage(ctx)).RoundInt()
	packetAmount := token.Amount.Sub(feeAmount)
	feeCoins := sdk.Coins{sdk.NewCoin(token.Denom, feeAmount)}
	packetCoin := sdk.NewCoin(token.Denom, packetAmount)

	// pay fees
	if feeAmount.IsPositive() {
		hostAccAddr, err := sdk.AccAddressFromBech32(receiver)
		if err != nil {
			return err
		}
		err = k.distrKeeper.FundCommunityPool(ctx, feeCoins, hostAccAddr)
		if err != nil {
			k.Logger(ctx).Error("packetForwardMiddleware error funding community pool",
				"error", err,
			)
			return errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
		}
	}

	memo := ""

	// set memo for next transfer with next from this transfer.
	if metadata.Next != nil {
		memoBz, err := json.Marshal(metadata.Next)
		if err != nil {
			k.Logger(ctx).Error("packetForwardMiddleware error marshaling next as JSON",
				"error", err,
			)
			return errorsmod.Wrapf(sdkerrors.ErrJSONMarshal, err.Error())
		}
		memo = string(memoBz)
	}

	msgTransfer := transfertypes.NewMsgTransfer(
		metadata.Port,
		metadata.Channel,
		packetCoin,
		receiver,
		metadata.Receiver,
		DefaultTransferPacketTimeoutHeight,
		uint64(ctx.BlockTime().UnixNano())+uint64(timeout.Nanoseconds()),
		memo,
	)

	k.Logger(ctx).Debug("packetForwardMiddleware ForwardTransferPacket",
		"port", metadata.Port, "channel", metadata.Channel,
		"sender", receiver, "receiver", metadata.Receiver,
		"amount", packetCoin.Amount.String(), "denom", packetCoin.Denom,
	)

	// send tokens to destination
	res, err := k.transferKeeper.Transfer(
		sdk.WrapSDKContext(ctx),
		msgTransfer,
	)
	if err != nil {
		k.Logger(ctx).Error("packetForwardMiddleware ForwardTransferPacket error",
			"port", metadata.Port, "channel", metadata.Channel,
			"sender", receiver, "receiver", metadata.Receiver,
			"amount", packetCoin.Amount.String(), "denom", packetCoin.Denom,
			"error", err,
		)
		return errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	// Store the following information in keeper:
	// key - information about forwarded packet: src_channel (parsedReceiver.Channel), src_port (parsedReceiver.Port), sequence
	// value - information about original packet for refunding if necessary: retries, srcPacketSender, srcPacket.DestinationChannel, srcPacket.DestinationPort

	if inFlightPacket == nil {
		inFlightPacket = &types.InFlightPacket{
			PacketData:            srcPacket.Data,
			OriginalSenderAddress: srcPacketSender,
			RefundChannelId:       srcPacket.DestinationChannel,
			RefundPortId:          srcPacket.DestinationPort,
			RefundSequence:        srcPacket.Sequence,
			PacketSrcPortId:       srcPacket.SourcePort,
			PacketSrcChannelId:    srcPacket.SourceChannel,

			PacketTimeoutTimestamp: srcPacket.TimeoutTimestamp,
			PacketTimeoutHeight:    srcPacket.TimeoutHeight.String(),

			RetriesRemaining: int32(maxRetries),
			Timeout:          uint64(timeout.Nanoseconds()),
			Nonrefundable:    nonrefundable,
		}
	} else {
		inFlightPacket.RetriesRemaining--
	}

	key := types.RefundPacketKey(metadata.Channel, metadata.Port, res.Sequence)
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(inFlightPacket)
	store.Set(key, bz)

	defer func() {
		if token.Amount.IsInt64() {
			telemetry.SetGaugeWithLabels(
				[]string{"tx", "msg", "ibc", "transfer"},
				float32(token.Amount.Int64()),
				[]metrics.Label{telemetry.NewLabel(coretypes.LabelDenom, token.Denom)},
			)
		}

		telemetry.IncrCounterWithLabels(
			[]string{"ibc", types.ModuleName, "send"},
			1,
			labels,
		)
	}()
	return nil
}

// TimeoutShouldRetry returns inFlightPacket and no error if retry should be attempted. Error is returned if IBC refund should occur.
func (k *Keeper) TimeoutShouldRetry(
	ctx sdk.Context,
	packet channeltypes.Packet,
) (*types.InFlightPacket, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.RefundPacketKey(packet.SourceChannel, packet.SourcePort, packet.Sequence)

	if !store.Has(key) {
		// not a forwarded packet, ignore.
		return nil, nil
	}

	bz := store.Get(key)
	var inFlightPacket types.InFlightPacket
	k.cdc.MustUnmarshal(bz, &inFlightPacket)

	if inFlightPacket.RetriesRemaining <= 0 {
		k.Logger(ctx).Error("packetForwardMiddleware reached max retries for packet",
			"key", string(key),
			"original-sender-address", inFlightPacket.OriginalSenderAddress,
			"refund-channel-id", inFlightPacket.RefundChannelId,
			"refund-port-id", inFlightPacket.RefundPortId,
		)
		return &inFlightPacket, fmt.Errorf("giving up on packet on channel (%s) port (%s) after max retries",
			inFlightPacket.RefundChannelId, inFlightPacket.RefundPortId)
	}

	return &inFlightPacket, nil
}

func (k *Keeper) RetryTimeout(
	ctx sdk.Context,
	channel, port string,
	data transfertypes.FungibleTokenPacketData,
	inFlightPacket *types.InFlightPacket,
) error {
	// send transfer again
	metadata := &types.ForwardMetadata{
		Receiver: data.Receiver,
		Channel:  channel,
		Port:     port,
	}

	if data.Memo != "" {
		metadata.Next = &types.JSONObject{}
		if err := json.Unmarshal([]byte(data.Memo), metadata.Next); err != nil {
			return fmt.Errorf("error unmarshaling memo json: %w", err)
		}
	}

	amount, ok := sdk.NewIntFromString(data.Amount)
	if !ok {
		k.Logger(ctx).Error("packetForwardMiddleware error parsing amount from string for router retry on timeout",
			"original-sender-address", inFlightPacket.OriginalSenderAddress,
			"refund-channel-id", inFlightPacket.RefundChannelId,
			"refund-port-id", inFlightPacket.RefundPortId,
			"retries-remaining", inFlightPacket.RetriesRemaining,
			"amount", data.Amount,
		)
		return fmt.Errorf("error parsing amount from string for router retry: %s", data.Amount)
	}

	denom := transfertypes.ParseDenomTrace(data.Denom).IBCDenom()

	token := sdk.NewCoin(denom, amount)

	// srcPacket and srcPacketSender are empty because inFlightPacket is non-nil.
	return k.ForwardTransferPacket(
		ctx,
		inFlightPacket,
		channeltypes.Packet{},
		"",
		data.Sender,
		metadata,
		token,
		uint8(inFlightPacket.RetriesRemaining),
		time.Duration(inFlightPacket.Timeout)*time.Nanosecond,
		nil,
		inFlightPacket.Nonrefundable,
	)
}

func (k *Keeper) RemoveInFlightPacket(ctx sdk.Context, packet channeltypes.Packet) {
	store := ctx.KVStore(k.storeKey)
	key := types.RefundPacketKey(packet.SourceChannel, packet.SourcePort, packet.Sequence)
	if !store.Has(key) {
		// not a forwarded packet, ignore.
		return
	}

	// done with packet key now, delete.
	store.Delete(key)
}

// GetAndClearInFlightPacket will fetch an InFlightPacket from the store, remove it if it exists, and return it.
func (k *Keeper) GetAndClearInFlightPacket(
	ctx sdk.Context,
	channel string,
	port string,
	sequence uint64,
) *types.InFlightPacket {
	store := ctx.KVStore(k.storeKey)
	key := types.RefundPacketKey(channel, port, sequence)
	if !store.Has(key) {
		// this is either not a forwarded packet, or it is the final destination for the refund.
		return nil
	}

	bz := store.Get(key)

	// done with packet key now, delete.
	store.Delete(key)

	var inFlightPacket types.InFlightPacket
	k.cdc.MustUnmarshal(bz, &inFlightPacket)
	return &inFlightPacket
}

// SendPacket wraps IBC ChannelKeeper's SendPacket function
func (k Keeper) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string, sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	return k.ics4Wrapper.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
}

// WriteAcknowledgement wraps IBC ICS4Wrapper WriteAcknowledgement function.
// ICS29 WriteAcknowledgement is used for asynchronous acknowledgements.
func (k *Keeper) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI, acknowledgement ibcexported.Acknowledgement) error {
	return k.ics4Wrapper.WriteAcknowledgement(ctx, chanCap, packet, acknowledgement)
}

// WriteAcknowledgement wraps IBC ICS4Wrapper GetAppVersion function.
func (k *Keeper) GetAppVersion(
	ctx sdk.Context,
	portID,
	channelID string,
) (string, bool) {
	return k.ics4Wrapper.GetAppVersion(ctx, portID, channelID)
}

// LookupModuleByChannel wraps ChannelKeeper LookupModuleByChannel function.
func (k *Keeper) LookupModuleByChannel(ctx sdk.Context, portID, channelID string) (string, *capabilitytypes.Capability, error) {
	return k.channelKeeper.LookupModuleByChannel(ctx, portID, channelID)
}
