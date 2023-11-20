package keeper

import (
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/testing/simapp/x/dummyware/types"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	"github.com/cometbft/cometbft/libs/log"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
)

// Keeper defines the packet forward middleware keeper
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	transferKeeper types.TransferKeeper
	channelKeeper  types.ChannelKeeper
	ics4Wrapper    porttypes.ICS4Wrapper
}

// NewKeeper creates a new forward Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	transferKeeper types.TransferKeeper,
	channelKeeper types.ChannelKeeper,
	ics4Wrapper porttypes.ICS4Wrapper,
) *Keeper {
	return &Keeper{
		cdc:            cdc,
		storeKey:       key,
		transferKeeper: transferKeeper,
		channelKeeper:  channelKeeper,
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

func (k *Keeper) WriteAcknowledgementForForcedNonRefundablePacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data transfertypes.FungibleTokenPacketData,
	inFlightPacket *types.InFlightPacket,
	ack channeltypes.Acknowledgement,
) error {
	inFlightPacket.Nonrefundable = true

	// Lookup module by channel capability
	_, chanCap, err := k.channelKeeper.LookupModuleByChannel(ctx, inFlightPacket.RefundPortId, inFlightPacket.RefundChannelId)
	if err != nil {
		return errorsmod.Wrap(err, "could not retrieve module from port-id")
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
