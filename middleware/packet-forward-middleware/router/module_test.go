package router_test

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/router"
	"testing"

	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/router/keeper"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/router/types"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/test"
	"github.com/golang/mock/gomock"
	"github.com/iancoleman/orderedmap"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

var (
	testDenom     = "uatom"
	testAmount    = "100"
	testAmount256 = "100000000000000000000"

	testSender             = "cosmos1xccz356ckk4xdx8nn7w9jlyaet2n98yzdkygse"
	testIntermediateAddr   = "cosmos1ye2e3njcuauzw7ccc62mmxh0j8nuslyqmrse2f8qg7tehtrxr8rqz2meck"
	testSourcePort         = "transfer"
	testSourceChannel      = "channel-10"
	testDestinationPort    = "transfer"
	testDestinationChannel = "channel-11"
)

func makeIBCDenom(port, channel, denom string) string {
	prefixedDenom := transfertypes.GetDenomPrefix(port, channel) + denom
	return transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()
}

func emptyPacket() channeltypes.Packet {
	return channeltypes.Packet{}
}

func transferPacket(t *testing.T, receiver string, metadata any) channeltypes.Packet {
	t.Helper()
	transferPacket := transfertypes.FungibleTokenPacketData{
		Sender:   testSender,
		Denom:    testDenom,
		Amount:   testAmount,
		Receiver: receiver,
	}

	if metadata != nil {
		if mStr, ok := metadata.(string); ok {
			transferPacket.Memo = mStr
		} else {
			memo, err := json.Marshal(metadata)
			require.NoError(t, err)
			transferPacket.Memo = string(memo)
		}
	}

	transferData, err := transfertypes.ModuleCdc.MarshalJSON(&transferPacket)
	require.NoError(t, err)

	return channeltypes.Packet{
		SourcePort:         testSourcePort,
		SourceChannel:      testSourceChannel,
		DestinationPort:    testDestinationPort,
		DestinationChannel: testDestinationChannel,
		Data:               transferData,
	}
}

func transferPacket256(t *testing.T, receiver string, metadata any) channeltypes.Packet {
	t.Helper()
	transferPacket := transfertypes.FungibleTokenPacketData{
		Sender:   testSender,
		Denom:    testDenom,
		Amount:   testAmount256,
		Receiver: receiver,
	}

	if metadata != nil {
		if mStr, ok := metadata.(string); ok {
			transferPacket.Memo = mStr
		} else {
			memo, err := json.Marshal(metadata)
			require.NoError(t, err)
			transferPacket.Memo = string(memo)
		}
	}

	transferData, err := transfertypes.ModuleCdc.MarshalJSON(&transferPacket)
	require.NoError(t, err)

	return channeltypes.Packet{
		SourcePort:         testSourcePort,
		SourceChannel:      testSourceChannel,
		DestinationPort:    testDestinationPort,
		DestinationChannel: testDestinationChannel,
		Data:               transferData,
	}
}

func TestOnRecvPacket_EmptyPacket(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	setup := test.NewTestSetup(t, ctl)
	ctx := setup.Initializer.Ctx
	cdc := setup.Initializer.Marshaler
	forwardMiddleware := setup.ForwardMiddleware

	// Test data
	senderAccAddr := test.AccAddress()
	packet := emptyPacket()

	ack := forwardMiddleware.OnRecvPacket(ctx, packet, senderAccAddr)
	require.False(t, ack.Success())

	expectedAck := &channeltypes.Acknowledgement{}
	err := cdc.UnmarshalJSON(ack.Acknowledgement(), expectedAck)
	require.NoError(t, err)
	require.Equal(t, "ABCI code: 1: error handling packet: see events for details", expectedAck.GetError())
}

func TestOnRecvPacket_InvalidReceiver(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	setup := test.NewTestSetup(t, ctl)
	ctx := setup.Initializer.Ctx
	cdc := setup.Initializer.Marshaler
	forwardMiddleware := setup.ForwardMiddleware

	// Test data
	senderAccAddr := test.AccAddress()
	packet := transferPacket(t, "", nil)

	// Expected mocks
	gomock.InOrder(
		setup.Mocks.IBCModuleMock.EXPECT().OnRecvPacket(ctx, packet, senderAccAddr).
			Return(channeltypes.NewResultAcknowledgement([]byte("test"))),
	)

	ack := forwardMiddleware.OnRecvPacket(ctx, packet, senderAccAddr)
	require.True(t, ack.Success())

	expectedAck := &channeltypes.Acknowledgement{}
	err := cdc.UnmarshalJSON(ack.Acknowledgement(), expectedAck)
	require.NoError(t, err)
}

func TestOnRecvPacket_NoForward(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	setup := test.NewTestSetup(t, ctl)
	ctx := setup.Initializer.Ctx
	cdc := setup.Initializer.Marshaler
	forwardMiddleware := setup.ForwardMiddleware

	// Test data
	senderAccAddr := test.AccAddress()
	packet := transferPacket(t, "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k", nil)

	// Expected mocks
	gomock.InOrder(
		setup.Mocks.IBCModuleMock.EXPECT().OnRecvPacket(ctx, packet, senderAccAddr).
			Return(channeltypes.NewResultAcknowledgement([]byte("test"))),
	)

	ack := forwardMiddleware.OnRecvPacket(ctx, packet, senderAccAddr)
	require.True(t, ack.Success())

	expectedAck := &channeltypes.Acknowledgement{}
	err := cdc.UnmarshalJSON(ack.Acknowledgement(), expectedAck)
	require.NoError(t, err)
	require.Equal(t, "test", string(expectedAck.GetResult()))
}

func TestOnRecvPacket_NoMemo(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	setup := test.NewTestSetup(t, ctl)
	ctx := setup.Initializer.Ctx
	cdc := setup.Initializer.Marshaler
	forwardMiddleware := setup.ForwardMiddleware

	// Test data
	senderAccAddr := test.AccAddress()
	packet := transferPacket(t, "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k", "{}")

	// Expected mocks
	gomock.InOrder(
		setup.Mocks.IBCModuleMock.EXPECT().OnRecvPacket(ctx, packet, senderAccAddr).
			Return(channeltypes.NewResultAcknowledgement([]byte("test"))),
	)

	ack := forwardMiddleware.OnRecvPacket(ctx, packet, senderAccAddr)
	require.True(t, ack.Success())

	expectedAck := &channeltypes.Acknowledgement{}
	err := cdc.UnmarshalJSON(ack.Acknowledgement(), expectedAck)
	require.NoError(t, err)
	require.Equal(t, "test", string(expectedAck.GetResult()))
}

func TestOnRecvPacket_RecvPacketFailed(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	setup := test.NewTestSetup(t, ctl)
	ctx := setup.Initializer.Ctx
	cdc := setup.Initializer.Marshaler
	forwardMiddleware := setup.ForwardMiddleware

	senderAccAddr := test.AccAddress()
	packet := transferPacket(t, "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k", nil)

	// Expected mocks
	gomock.InOrder(
		// We return a failed OnRecvPacket
		setup.Mocks.IBCModuleMock.EXPECT().OnRecvPacket(ctx, packet, senderAccAddr).
			Return(channeltypes.NewErrorAcknowledgement(fmt.Errorf("test"))),
	)

	ack := forwardMiddleware.OnRecvPacket(ctx, packet, senderAccAddr)
	require.False(t, ack.Success())

	expectedAck := &channeltypes.Acknowledgement{}
	err := cdc.UnmarshalJSON(ack.Acknowledgement(), expectedAck)
	require.NoError(t, err)
	require.Equal(t, "ABCI code: 1: error handling packet: see events for details", expectedAck.GetError())
}

func TestOnRecvPacket_ForwardNoFee(t *testing.T) {
	var err error
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	setup := test.NewTestSetup(t, ctl)
	ctx := setup.Initializer.Ctx
	cdc := setup.Initializer.Marshaler
	forwardMiddleware := setup.ForwardMiddleware

	// Test data
	const (
		hostAddr = "cosmos1vzxkv3lxccnttr9rs0002s93sgw72h7ghukuhs"
		destAddr = "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k"
		port     = "transfer"
		channel  = "channel-0"
	)
	denom := makeIBCDenom(testDestinationPort, testDestinationChannel, testDenom)
	senderAccAddr := test.AccAddress()
	testCoin := sdk.NewCoin(denom, sdk.NewInt(100))
	packetOrig := transferPacket(t, hostAddr, &types.PacketMetadata{
		Forward: &types.ForwardMetadata{
			Receiver: destAddr,
			Port:     port,
			Channel:  channel,
		},
	})
	packetOrigWithNewReceiver := transferPacket(t, testIntermediateAddr, &types.PacketMetadata{
		Forward: &types.ForwardMetadata{
			Receiver: destAddr,
			Port:     port,
			Channel:  channel,
		},
	})
	packetFwd := transferPacket(t, destAddr, nil)

	acknowledgement := channeltypes.NewResultAcknowledgement([]byte("test"))
	successAck := cdc.MustMarshalJSON(&acknowledgement)

	// Expected mocks
	gomock.InOrder(
		setup.Mocks.IBCModuleMock.EXPECT().OnRecvPacket(ctx, packetOrigWithNewReceiver, senderAccAddr).
			Return(acknowledgement),

		setup.Mocks.TransferKeeperMock.EXPECT().Transfer(
			sdk.WrapSDKContext(ctx),
			transfertypes.NewMsgTransfer(
				port,
				channel,
				testCoin,
				testIntermediateAddr,
				destAddr,
				keeper.DefaultTransferPacketTimeoutHeight,
				uint64(ctx.BlockTime().UnixNano())+uint64(keeper.DefaultForwardTransferPacketTimeoutTimestamp.Nanoseconds()),
				"",
			),
		).Return(&transfertypes.MsgTransferResponse{Sequence: 0}, nil),

		setup.Mocks.IBCModuleMock.EXPECT().OnAcknowledgementPacket(ctx, packetFwd, successAck, senderAccAddr).
			Return(nil),
	)

	// chain B with router module receives packet and forwards. ack should be nil so that it is not written yet.
	ack := forwardMiddleware.OnRecvPacket(ctx, packetOrig, senderAccAddr)
	require.Nil(t, ack)

	// ack returned from chain C
	err = forwardMiddleware.OnAcknowledgementPacket(ctx, packetFwd, successAck, senderAccAddr)
	require.NoError(t, err)
}

func TestOnRecvPacket_ForwardAmountInt256(t *testing.T) {
	var err error
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	setup := test.NewTestSetup(t, ctl)
	ctx := setup.Initializer.Ctx
	cdc := setup.Initializer.Marshaler
	forwardMiddleware := setup.ForwardMiddleware

	// Test data
	const (
		hostAddr = "cosmos1vzxkv3lxccnttr9rs0002s93sgw72h7ghukuhs"
		destAddr = "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k"
		port     = "transfer"
		channel  = "channel-0"
	)
	denom := makeIBCDenom(testDestinationPort, testDestinationChannel, testDenom)
	senderAccAddr := test.AccAddress()

	amount256, ok := sdk.NewIntFromString(testAmount256)
	require.True(t, ok)

	testCoin := sdk.NewCoin(denom, amount256)
	packetOrig := transferPacket256(t, hostAddr, &types.PacketMetadata{
		Forward: &types.ForwardMetadata{
			Receiver: destAddr,
			Port:     port,
			Channel:  channel,
		},
	})
	packetOrigWithNewReceiver := transferPacket256(t, testIntermediateAddr, &types.PacketMetadata{
		Forward: &types.ForwardMetadata{
			Receiver: destAddr,
			Port:     port,
			Channel:  channel,
		},
	})
	packetFwd := transferPacket256(t, destAddr, nil)

	acknowledgement := channeltypes.NewResultAcknowledgement([]byte("test"))
	successAck := cdc.MustMarshalJSON(&acknowledgement)

	// Expected mocks
	gomock.InOrder(
		setup.Mocks.IBCModuleMock.EXPECT().OnRecvPacket(ctx, packetOrigWithNewReceiver, senderAccAddr).
			Return(acknowledgement),

		setup.Mocks.TransferKeeperMock.EXPECT().Transfer(
			sdk.WrapSDKContext(ctx),
			transfertypes.NewMsgTransfer(
				port,
				channel,
				testCoin,
				testIntermediateAddr,
				destAddr,
				keeper.DefaultTransferPacketTimeoutHeight,
				uint64(ctx.BlockTime().UnixNano())+uint64(keeper.DefaultForwardTransferPacketTimeoutTimestamp.Nanoseconds()),
				"",
			),
		).Return(&transfertypes.MsgTransferResponse{Sequence: 0}, nil),

		setup.Mocks.IBCModuleMock.EXPECT().OnAcknowledgementPacket(ctx, packetFwd, successAck, senderAccAddr).
			Return(nil),
	)

	// chain B with router module receives packet and forwards. ack should be nil so that it is not written yet.
	ack := forwardMiddleware.OnRecvPacket(ctx, packetOrig, senderAccAddr)
	require.Nil(t, ack)

	// ack returned from chain C
	err = forwardMiddleware.OnAcknowledgementPacket(ctx, packetFwd, successAck, senderAccAddr)
	require.NoError(t, err)
}

func TestOnRecvPacket_ForwardWithFee(t *testing.T) {
	var err error
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	setup := test.NewTestSetup(t, ctl)
	ctx := setup.Initializer.Ctx
	cdc := setup.Initializer.Marshaler
	forwardMiddleware := setup.ForwardMiddleware

	// Set fee param to 10%
	setup.Keepers.RouterKeeper.SetParams(ctx, types.NewParams(sdk.NewDecWithPrec(10, 2)))

	// Test data
	const (
		hostAddr = "cosmos1vzxkv3lxccnttr9rs0002s93sgw72h7ghukuhs"
		destAddr = "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k"
		port     = "transfer"
		channel  = "channel-0"
	)
	denom := makeIBCDenom(testDestinationPort, testDestinationChannel, testDenom)
	senderAccAddr := test.AccAddress()
	intermediateAccAddr := test.AccAddressFromBech32(t, testIntermediateAddr)
	testCoin := sdk.NewCoin(denom, sdk.NewInt(90))
	feeCoins := sdk.Coins{sdk.NewCoin(denom, sdk.NewInt(10))}
	packetOrig := transferPacket(t, hostAddr, &types.PacketMetadata{
		Forward: &types.ForwardMetadata{
			Receiver: destAddr,
			Port:     port,
			Channel:  channel,
		},
	})
	packetOrigWithNewReceiver := transferPacket(t, testIntermediateAddr, &types.PacketMetadata{
		Forward: &types.ForwardMetadata{
			Receiver: destAddr,
			Port:     port,
			Channel:  channel,
		},
	})
	packetFwd := transferPacket(t, destAddr, nil)
	acknowledgement := channeltypes.NewResultAcknowledgement([]byte("test"))
	successAck := cdc.MustMarshalJSON(&acknowledgement)

	// Expected mocks
	gomock.InOrder(
		setup.Mocks.IBCModuleMock.EXPECT().OnRecvPacket(ctx, packetOrigWithNewReceiver, senderAccAddr).
			Return(acknowledgement),

		setup.Mocks.DistributionKeeperMock.EXPECT().FundCommunityPool(
			ctx,
			feeCoins,
			intermediateAccAddr,
		).Return(nil),

		setup.Mocks.TransferKeeperMock.EXPECT().Transfer(
			sdk.WrapSDKContext(ctx),
			transfertypes.NewMsgTransfer(
				port,
				channel,
				testCoin,
				testIntermediateAddr,
				destAddr,
				keeper.DefaultTransferPacketTimeoutHeight,
				uint64(ctx.BlockTime().UnixNano())+uint64(keeper.DefaultForwardTransferPacketTimeoutTimestamp.Nanoseconds()),
				"",
			),
		).Return(&transfertypes.MsgTransferResponse{Sequence: 0}, nil),

		setup.Mocks.IBCModuleMock.EXPECT().OnAcknowledgementPacket(ctx, packetFwd, successAck, senderAccAddr).
			Return(nil),
	)

	// chain B with router module receives packet and forwards. ack should be nil so that it is not written yet.
	ack := forwardMiddleware.OnRecvPacket(ctx, packetOrig, senderAccAddr)
	require.Nil(t, ack)

	// ack returned from chain C
	err = forwardMiddleware.OnAcknowledgementPacket(ctx, packetFwd, successAck, senderAccAddr)
	require.NoError(t, err)
}

func TestOnRecvPacket_ForwardMultihopStringNext(t *testing.T) {
	var err error
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	setup := test.NewTestSetup(t, ctl)
	ctx := setup.Initializer.Ctx
	cdc := setup.Initializer.Marshaler
	forwardMiddleware := setup.ForwardMiddleware

	// Test data
	const (
		hostAddr  = "cosmos1vzxkv3lxccnttr9rs0002s93sgw72h7ghukuhs"
		hostAddr2 = "cosmos1q4p4gx889lfek5augdurrjclwtqvjhuntm6j4m"
		destAddr  = "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k"
		port      = "transfer"
		channel   = "channel-0"
		channel2  = "channel-1"
	)

	denom := makeIBCDenom(testDestinationPort, testDestinationChannel, testDenom)
	senderAccAddr := test.AccAddress()
	senderAccAddr2 := test.AccAddress()
	testCoin := sdk.NewCoin(denom, sdk.NewInt(100))
	nextMetadata := &types.PacketMetadata{
		Forward: &types.ForwardMetadata{
			Receiver: destAddr,
			Port:     port,
			Channel:  channel2,
		},
	}
	nextBz, err := json.Marshal(nextMetadata)
	require.NoError(t, err)

	packetOrig := transferPacket(t, hostAddr, &types.PacketMetadata{
		Forward: &types.ForwardMetadata{
			Receiver: hostAddr2,
			Port:     port,
			Channel:  channel,
			Next:     types.NewJSONObject(false, nextBz, orderedmap.OrderedMap{}),
		},
	})

	packetOrigWithNewReceiver := transferPacket(t, testIntermediateAddr, &types.PacketMetadata{
		Forward: &types.ForwardMetadata{
			Receiver: hostAddr2,
			Port:     port,
			Channel:  channel,
			Next:     types.NewJSONObject(false, nextBz, orderedmap.OrderedMap{}),
		},
	})

	packet2 := transferPacket(t, hostAddr2, nextMetadata)
	packet2WithNewReceiver := transferPacket(t, testIntermediateAddr, nextMetadata)
	packetFwd := transferPacket(t, destAddr, nil)

	memo1, err := json.Marshal(nextMetadata)
	require.NoError(t, err)

	msgTransfer1 := transfertypes.NewMsgTransfer(
		port,
		channel,
		testCoin,
		testIntermediateAddr,
		hostAddr2,
		keeper.DefaultTransferPacketTimeoutHeight,
		uint64(ctx.BlockTime().UnixNano())+uint64(keeper.DefaultForwardTransferPacketTimeoutTimestamp.Nanoseconds()),
		string(memo1),
	)

	// no memo on final forward
	msgTransfer2 := transfertypes.NewMsgTransfer(
		port,
		channel2,
		testCoin,
		testIntermediateAddr,
		destAddr,
		keeper.DefaultTransferPacketTimeoutHeight,
		uint64(ctx.BlockTime().UnixNano())+uint64(keeper.DefaultForwardTransferPacketTimeoutTimestamp.Nanoseconds()),
		"",
	)

	acknowledgement := channeltypes.NewResultAcknowledgement([]byte("test"))
	successAck := cdc.MustMarshalJSON(&acknowledgement)

	// Expected mocks
	gomock.InOrder(
		setup.Mocks.IBCModuleMock.EXPECT().OnRecvPacket(ctx, packetOrigWithNewReceiver, senderAccAddr).
			Return(acknowledgement),

		setup.Mocks.TransferKeeperMock.EXPECT().Transfer(
			sdk.WrapSDKContext(ctx),
			msgTransfer1,
		).Return(&transfertypes.MsgTransferResponse{Sequence: 0}, nil),

		setup.Mocks.IBCModuleMock.EXPECT().OnRecvPacket(ctx, packet2WithNewReceiver, senderAccAddr2).
			Return(acknowledgement),

		setup.Mocks.TransferKeeperMock.EXPECT().Transfer(
			sdk.WrapSDKContext(ctx),
			msgTransfer2,
		).Return(&transfertypes.MsgTransferResponse{Sequence: 0}, nil),

		setup.Mocks.IBCModuleMock.EXPECT().OnAcknowledgementPacket(ctx, packetFwd, successAck, senderAccAddr2).
			Return(nil),

		setup.Mocks.IBCModuleMock.EXPECT().OnAcknowledgementPacket(ctx, packet2, successAck, senderAccAddr).
			Return(nil),
	)

	// chain B with router module receives packet and forwards. ack should be nil so that it is not written yet.
	ack := forwardMiddleware.OnRecvPacket(ctx, packetOrig, senderAccAddr)
	require.Nil(t, ack)

	// chain C with router module receives packet and forwards. ack should be nil so that it is not written yet.
	ack = forwardMiddleware.OnRecvPacket(ctx, packet2, senderAccAddr2)
	require.Nil(t, ack)

	// ack returned from chain D to chain C
	err = forwardMiddleware.OnAcknowledgementPacket(ctx, packetFwd, successAck, senderAccAddr2)
	require.NoError(t, err)

	// ack returned from chain C to chain B
	err = forwardMiddleware.OnAcknowledgementPacket(ctx, packet2, successAck, senderAccAddr)
	require.NoError(t, err)
}

func TestOnRecvPacket_ForwardMultihopJSONNext(t *testing.T) {
	var err error
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	setup := test.NewTestSetup(t, ctl)
	ctx := setup.Initializer.Ctx
	cdc := setup.Initializer.Marshaler
	forwardMiddleware := setup.ForwardMiddleware

	// Test data
	const (
		hostAddr  = "cosmos1vzxkv3lxccnttr9rs0002s93sgw72h7ghukuhs"
		hostAddr2 = "cosmos1q4p4gx889lfek5augdurrjclwtqvjhuntm6j4m"
		destAddr  = "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k"
		port      = "transfer"
		channel   = "channel-0"
		channel2  = "channel-1"
	)

	denom := makeIBCDenom(testDestinationPort, testDestinationChannel, testDenom)
	senderAccAddr := test.AccAddress()
	senderAccAddr2 := test.AccAddress()
	testCoin := sdk.NewCoin(denom, sdk.NewInt(100))
	nextMetadata := &types.PacketMetadata{
		Forward: &types.ForwardMetadata{
			Receiver: destAddr,
			Port:     port,
			Channel:  channel2,
		},
	}
	nextBz, err := json.Marshal(nextMetadata)
	require.NoError(t, err)

	nextJSONObject := new(types.JSONObject)
	err = json.Unmarshal(nextBz, nextJSONObject)
	require.NoError(t, err)

	packetOrig := transferPacket(t, hostAddr, &types.PacketMetadata{
		Forward: &types.ForwardMetadata{
			Receiver: hostAddr2,
			Port:     port,
			Channel:  channel,
			Next:     nextJSONObject,
		},
	})

	packetOrigWithNewReceiver := transferPacket(t, testIntermediateAddr, &types.PacketMetadata{
		Forward: &types.ForwardMetadata{
			Receiver: hostAddr2,
			Port:     port,
			Channel:  channel,
			Next:     types.NewJSONObject(false, nextBz, orderedmap.OrderedMap{}),
		},
	})

	packet2 := transferPacket(t, hostAddr2, nextMetadata)
	packet2WithNewReceiver := transferPacket(t, testIntermediateAddr, nextMetadata)
	packetFwd := transferPacket(t, destAddr, nil)

	memo1, err := json.Marshal(nextMetadata)
	require.NoError(t, err)

	msgTransfer1 := transfertypes.NewMsgTransfer(
		port,
		channel,
		testCoin,
		testIntermediateAddr,
		hostAddr2,
		keeper.DefaultTransferPacketTimeoutHeight,
		uint64(ctx.BlockTime().UnixNano())+uint64(keeper.DefaultForwardTransferPacketTimeoutTimestamp.Nanoseconds()),
		string(memo1),
	)

	// no memo on final forward
	msgTransfer2 := transfertypes.NewMsgTransfer(
		port,
		channel2,
		testCoin,
		testIntermediateAddr,
		destAddr,
		keeper.DefaultTransferPacketTimeoutHeight,
		uint64(ctx.BlockTime().UnixNano())+uint64(keeper.DefaultForwardTransferPacketTimeoutTimestamp.Nanoseconds()),
		"",
	)

	acknowledgement := channeltypes.NewResultAcknowledgement([]byte("test"))
	successAck := cdc.MustMarshalJSON(&acknowledgement)

	// Expected mocks
	gomock.InOrder(
		setup.Mocks.IBCModuleMock.EXPECT().OnRecvPacket(ctx, packetOrigWithNewReceiver, senderAccAddr).
			Return(acknowledgement),

		setup.Mocks.TransferKeeperMock.EXPECT().Transfer(
			sdk.WrapSDKContext(ctx),
			msgTransfer1,
		).Return(&transfertypes.MsgTransferResponse{Sequence: 0}, nil),

		setup.Mocks.IBCModuleMock.EXPECT().OnRecvPacket(ctx, packet2WithNewReceiver, senderAccAddr2).
			Return(acknowledgement),

		setup.Mocks.TransferKeeperMock.EXPECT().Transfer(
			sdk.WrapSDKContext(ctx),
			msgTransfer2,
		).Return(&transfertypes.MsgTransferResponse{Sequence: 0}, nil),

		setup.Mocks.IBCModuleMock.EXPECT().OnAcknowledgementPacket(ctx, packetFwd, successAck, senderAccAddr2).
			Return(nil),

		setup.Mocks.IBCModuleMock.EXPECT().OnAcknowledgementPacket(ctx, packet2, successAck, senderAccAddr).
			Return(nil),
	)

	// chain B with router module receives packet and forwards. ack should be nil so that it is not written yet.
	ack := forwardMiddleware.OnRecvPacket(ctx, packetOrig, senderAccAddr)
	require.Nil(t, ack)

	// chain C with router module receives packet and forwards. ack should be nil so that it is not written yet.
	ack = forwardMiddleware.OnRecvPacket(ctx, packet2, senderAccAddr2)
	require.Nil(t, ack)

	// ack returned from chain D to chain C
	err = forwardMiddleware.OnAcknowledgementPacket(ctx, packetFwd, successAck, senderAccAddr2)
	require.NoError(t, err)

	// ack returned from chain C to chain B
	err = forwardMiddleware.OnAcknowledgementPacket(ctx, packet2, successAck, senderAccAddr)
	require.NoError(t, err)
}

func TestDeriveIntermediateAddress(t *testing.T) {
	intermediateAddr, err := router.DeriveIntermediateAddress(testSourcePort, testSourceChannel, testSender)
	require.NoError(t, err)
	require.Equal(t, testIntermediateAddr, intermediateAddr)
}
