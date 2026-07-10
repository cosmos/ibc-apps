package ratelimit_test

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10"
	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10/keeper"
	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10/testing/simapp/apptesting"
	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
)

const (
	testTransferPort = "transfer"
	testSequence     = uint64(1)
	invalidPacketBz  = "invalid packet data"
)

// mockRecvModule implements porttypes.IBCModule for the packet callbacks used in
// these tests. The embedded interface panics for any other callback.
type mockRecvModule struct {
	porttypes.IBCModule

	recvAck ibcexported.Acknowledgement
}

func (m mockRecvModule) OnRecvPacket(sdk.Context, string, channeltypes.Packet, sdk.AccAddress) ibcexported.Acknowledgement {
	return m.recvAck
}

type mockICS4Wrapper struct {
	writeAckCalled bool
}

func (*mockICS4Wrapper) SendPacket(sdk.Context, string, string, clienttypes.Height, uint64, []byte) (uint64, error) {
	return 0, nil
}

func (m *mockICS4Wrapper) WriteAcknowledgement(sdk.Context, ibcexported.PacketI, ibcexported.Acknowledgement) error {
	m.writeAckCalled = true
	return nil
}

func (*mockICS4Wrapper) GetAppVersion(sdk.Context, string, string) (string, bool) {
	return "", false
}

func TestWriteAcknowledgement_NilAck(t *testing.T) {
	middleware := ratelimit.NewIBCMiddleware(keeper.Keeper{}, nil)
	packet := channeltypes.Packet{
		Sequence:           1,
		DestinationChannel: "channel-0",
	}

	var ack ibcexported.Acknowledgement
	err := middleware.WriteAcknowledgement(sdk.Context{}, packet, ack)

	require.ErrorIs(t, err, types.ErrAsyncAckNil)
	require.ErrorContains(t, err, "cannot write nil ack for packet channel-0/1")
}

func TestOnRecvPacketRemovesPendingReceivePacketForSyncAck(t *testing.T) {
	helper := apptesting.AppTestHelper{}
	helper.Setup()
	ctx := helper.Ctx
	ratelimitKeeper := helper.App.RatelimitKeeper

	packet, packetInfo := createRecvPacket(t)
	setReceiveRateLimit(ctx, ratelimitKeeper, packetInfo)

	expectedAck := channeltypes.NewResultAcknowledgement([]byte{1})
	middleware := ratelimit.NewIBCMiddleware(ratelimitKeeper, mockRecvModule{recvAck: expectedAck})

	ack := middleware.OnRecvPacket(ctx, "", packet, nil)
	require.Equal(t, expectedAck, ack)

	found, err := ratelimitKeeper.CheckPacketReceivedDuringCurrentQuota(ctx, packetInfo.ChannelID, packet.Sequence, packetInfo.Denom)
	require.NoError(t, err)
	require.False(t, found)
}

func TestWriteAcknowledgementRemovesPendingReceivePacketForSuccessAck(t *testing.T) {
	helper := apptesting.AppTestHelper{}
	helper.Setup()
	ctx := helper.Ctx
	ratelimitKeeper := helper.App.RatelimitKeeper
	writeAckWrapper := &mockICS4Wrapper{}
	ratelimitKeeper.SetIBCKeepers(nil, nil, writeAckWrapper)

	packet, packetInfo := createRecvPacket(t)
	require.NoError(t, ratelimitKeeper.SetPendingReceivePacket(ctx, packetInfo.ChannelID, packet.Sequence, packetInfo.Denom))

	middleware := ratelimit.NewIBCMiddleware(ratelimitKeeper, nil)
	ack := channeltypes.NewResultAcknowledgement([]byte{1})

	err := middleware.WriteAcknowledgement(ctx, packet, ack)
	require.NoError(t, err)
	require.True(t, writeAckWrapper.writeAckCalled)

	found, err := ratelimitKeeper.CheckPacketReceivedDuringCurrentQuota(ctx, packetInfo.ChannelID, packet.Sequence, packetInfo.Denom)
	require.NoError(t, err)
	require.False(t, found)
}

func TestWriteAcknowledgementFallsBackWhenPendingCleanupParseFails(t *testing.T) {
	helper := apptesting.AppTestHelper{}
	helper.Setup()
	ctx := helper.Ctx
	writeAckWrapper := &mockICS4Wrapper{}
	helper.App.RatelimitKeeper.SetIBCKeepers(nil, nil, writeAckWrapper)

	middleware := ratelimit.NewIBCMiddleware(helper.App.RatelimitKeeper, nil)
	ack := channeltypes.NewResultAcknowledgement([]byte{1})

	err := middleware.WriteAcknowledgement(ctx, createInvalidRecvPacket(), ack)
	require.NoError(t, err)
	require.True(t, writeAckWrapper.writeAckCalled)
}

func createRecvPacket(t *testing.T) (channeltypes.Packet, keeper.RateLimitedPacketInfo) {
	t.Helper()

	packetData := transfertypes.FungibleTokenPacketData{
		Denom:    "uosmo",
		Amount:   "10",
		Sender:   "sender",
		Receiver: "receiver",
	}
	packetDataBz, err := json.Marshal(packetData)
	require.NoError(t, err)

	packet := channeltypes.NewPacket(
		packetDataBz,
		testSequence,
		testTransferPort,
		"channel-1",
		testTransferPort,
		"channel-0",
		clienttypes.Height{},
		1,
	)
	packetInfo, err := keeper.ParsePacketInfo(packet, types.PACKET_RECV)
	require.NoError(t, err)

	return packet, packetInfo
}

func createInvalidRecvPacket() channeltypes.Packet {
	return channeltypes.NewPacket(
		[]byte(invalidPacketBz),
		testSequence,
		testTransferPort,
		"channel-1",
		testTransferPort,
		"channel-0",
		clienttypes.Height{},
		1,
	)
}

func setReceiveRateLimit(ctx sdk.Context, ratelimitKeeper keeper.Keeper, packetInfo keeper.RateLimitedPacketInfo) {
	ratelimitKeeper.SetRateLimit(ctx, types.RateLimit{
		Path: &types.Path{
			Denom:             packetInfo.Denom,
			ChannelOrClientId: packetInfo.ChannelID,
		},
		Quota: &types.Quota{
			MaxPercentSend: sdkmath.NewInt(100),
			MaxPercentRecv: sdkmath.NewInt(100),
		},
		Flow: &types.Flow{
			Inflow:       sdkmath.ZeroInt(),
			Outflow:      sdkmath.ZeroInt(),
			ChannelValue: sdkmath.NewInt(100),
		},
	})
}
