package v2_test

import (
	"errors"
	"testing"

	ratelimitkeeper "github.com/cosmos/ibc-apps/modules/rate-limiting/v10/keeper"
	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10/testing/simapp/apptesting"
	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"
	ratelimitv2 "github.com/cosmos/ibc-apps/modules/rate-limiting/v10/v2"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"
	channeltypesv2 "github.com/cosmos/ibc-go/v11/modules/core/04-channel/v2/types"
)

type mockIBCModule struct{}

func (mockIBCModule) OnSendPacket(sdk.Context, string, string, uint64, channeltypesv2.Payload, sdk.AccAddress) error {
	return nil
}

func (mockIBCModule) OnRecvPacket(sdk.Context, string, string, uint64, channeltypesv2.Payload, sdk.AccAddress) channeltypesv2.RecvPacketResult {
	return channeltypesv2.RecvPacketResult{}
}

func (mockIBCModule) OnTimeoutPacket(sdk.Context, string, string, uint64, channeltypesv2.Payload, sdk.AccAddress) error {
	return nil
}

func (mockIBCModule) OnAcknowledgementPacket(sdk.Context, string, string, uint64, []byte, channeltypesv2.Payload, sdk.AccAddress) error {
	return nil
}

type mockWriteAckWrapper struct {
	called  bool
	ack     channeltypesv2.Acknowledgement
	client  string
	seq     uint64
	callErr error
}

func (m *mockWriteAckWrapper) WriteAcknowledgement(_ sdk.Context, clientID string, sequence uint64, ack channeltypesv2.Acknowledgement) error {
	m.called = true
	m.client = clientID
	m.seq = sequence
	m.ack = ack
	return m.callErr
}

type mockChannelKeeperV2 struct {
	packet channeltypesv2.Packet
	found  bool
}

func (m mockChannelKeeperV2) GetAsyncPacket(sdk.Context, string, uint64) (channeltypesv2.Packet, bool) {
	return m.packet, m.found
}

func TestNewIBCMiddleware(t *testing.T) {
	testCases := []struct {
		name          string
		instantiateFn func()
		expectedPanic string
	}{
		{
			name: "success: full constructor",
			instantiateFn: func() {
				_ = ratelimitv2.NewIBCMiddlewareWithAsyncAcknowledgements(ratelimitkeeper.Keeper{}, mockIBCModule{}, &mockWriteAckWrapper{}, mockChannelKeeperV2{})
			},
		},
		{
			name: "success: deprecated constructor",
			instantiateFn: func() {
				_ = ratelimitv2.NewIBCMiddleware(ratelimitkeeper.Keeper{}, mockIBCModule{})
			},
		},
		{
			name: "failure: nil write acknowledgement wrapper in full constructor",
			instantiateFn: func() {
				_ = ratelimitv2.NewIBCMiddlewareWithAsyncAcknowledgements(ratelimitkeeper.Keeper{}, mockIBCModule{}, nil, mockChannelKeeperV2{})
			},
			expectedPanic: ratelimittypes.ErrWriteAcknowledgementWrapperNil.Error(),
		},
		{
			name: "failure: nil channel keeper v2 in full constructor",
			instantiateFn: func() {
				_ = ratelimitv2.NewIBCMiddlewareWithAsyncAcknowledgements(ratelimitkeeper.Keeper{}, mockIBCModule{}, &mockWriteAckWrapper{}, nil)
			},
			expectedPanic: ratelimittypes.ErrChannelKeeperV2Nil.Error(),
		},
		{
			name: "failure: nil write acknowledgement wrapper in setter",
			instantiateFn: func() {
				mw := ratelimitv2.NewIBCMiddleware(ratelimitkeeper.Keeper{}, mockIBCModule{})
				mw.SetWriteAcknowledgementWrapper(nil)
			},
			expectedPanic: ratelimittypes.ErrWriteAcknowledgementWrapperNil.Error(),
		},
		{
			name: "failure: nil channel keeper v2 in setter",
			instantiateFn: func() {
				mw := ratelimitv2.NewIBCMiddleware(ratelimitkeeper.Keeper{}, mockIBCModule{})
				mw.SetChannelKeeperV2(nil)
			},
			expectedPanic: ratelimittypes.ErrChannelKeeperV2Nil.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectedPanic == "" {
				require.NotPanics(t, tc.instantiateFn)
			} else {
				require.PanicsWithError(t, tc.expectedPanic, tc.instantiateFn)
			}
		})
	}
}

func TestWriteAcknowledgement_MissingDependencies(t *testing.T) {
	mw := ratelimitv2.NewIBCMiddleware(ratelimitkeeper.Keeper{}, mockIBCModule{})
	err := mw.WriteAcknowledgement(sdk.Context{}, "client", 1, channeltypesv2.NewAcknowledgement([]byte("success")))
	require.ErrorIs(t, err, ratelimittypes.ErrChannelKeeperV2Nil)

	mw.SetChannelKeeperV2(mockChannelKeeperV2{})
	err = mw.WriteAcknowledgement(sdk.Context{}, "client", 1, channeltypesv2.NewAcknowledgement([]byte("success")))
	require.ErrorIs(t, err, ratelimittypes.ErrWriteAcknowledgementWrapperNil)
}

func TestWriteAcknowledgement(t *testing.T) {
	const (
		sequence          = uint64(1)
		sourceClient      = "sourceClient"
		destinationClient = "destinationClient"
		uosmo             = "uosmo"
		transferPort      = "transfer"
	)

	packetAmount := sdkmath.NewInt(10)
	errorAck := channeltypesv2.NewAcknowledgement(channeltypesv2.ErrorAcknowledgement[:])
	successAck := channeltypesv2.NewAcknowledgement([]byte("success"))
	writeAckErr := "write acknowledgement failed"

	testCases := []struct {
		name            string
		ack             channeltypesv2.Acknowledgement
		asyncFound      bool
		malleatePayload func(*channeltypesv2.Payload)
		writeAckErr     error
		expectedErr     string
		expectWriteAck  bool
		checkInflow     bool
		expectedInflow  sdkmath.Int
	}{
		{
			name:           "success: error acknowledgement undoes receive inflow",
			ack:            errorAck,
			asyncFound:     true,
			expectWriteAck: true,
			checkInflow:    true,
			expectedInflow: sdkmath.NewInt(90),
		},
		{
			name:           "success: success acknowledgement does not undo receive inflow",
			ack:            successAck,
			asyncFound:     true,
			expectWriteAck: true,
			checkInflow:    true,
			expectedInflow: sdkmath.NewInt(100),
		},
		{
			name:           "failure: missing async packet",
			ack:            errorAck,
			checkInflow:    true,
			expectedInflow: sdkmath.NewInt(100),
			expectedErr:    "async packet not found",
		},
		{
			name:       "failure: async packet cannot be converted",
			ack:        errorAck,
			asyncFound: true,
			malleatePayload: func(payload *channeltypesv2.Payload) {
				payload.Encoding = "invalid"
				payload.Value = []byte("invalid packet data")
			},
			expectedErr: "invalid encoding",
		},
		{
			name:           "failure: write acknowledgement wrapper returns error",
			ack:            successAck,
			asyncFound:     true,
			writeAckErr:    errors.New(writeAckErr),
			expectedErr:    writeAckErr,
			expectWriteAck: true,
			checkInflow:    true,
			expectedInflow: sdkmath.NewInt(100),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			helper := apptesting.AppTestHelper{}
			helper.Setup()
			ctx := helper.Ctx

			packetData := transfertypes.FungibleTokenPacketData{
				Denom:    uosmo,
				Amount:   packetAmount.String(),
				Sender:   "sender",
				Receiver: "receiver",
			}
			packetDataBz, err := transfertypes.MarshalPacketData(packetData, transfertypes.V1, transfertypes.EncodingJSON)
			require.NoError(t, err)

			payload := channeltypesv2.Payload{
				SourcePort:      transferPort,
				DestinationPort: transferPort,
				Version:         transfertypes.V1,
				Encoding:        transfertypes.EncodingJSON,
				Value:           packetDataBz,
			}
			if tc.malleatePayload != nil {
				tc.malleatePayload(&payload)
			}

			packet := channeltypesv2.NewPacket(sequence, sourceClient, destinationClient, 0, payload)
			rateLimitDenom := transfertypes.ParseDenomTrace(transfertypes.GetDenomPrefix(transferPort, destinationClient) + uosmo).IBCDenom()
			if tc.checkInflow {
				helper.App.RatelimitKeeper.SetRateLimit(ctx, ratelimittypes.RateLimit{
					Path: &ratelimittypes.Path{Denom: rateLimitDenom, ChannelOrClientId: destinationClient},
					Flow: &ratelimittypes.Flow{Inflow: sdkmath.NewInt(100)},
				})
				if tc.asyncFound {
					err = helper.App.RatelimitKeeper.SetPendingReceivePacket(ctx, destinationClient, sequence, rateLimitDenom)
					require.NoError(t, err)
				}
			}

			writeAckWrapper := &mockWriteAckWrapper{callErr: tc.writeAckErr}
			mw := ratelimitv2.NewIBCMiddlewareWithAsyncAcknowledgements(
				helper.App.RatelimitKeeper,
				mockIBCModule{},
				writeAckWrapper,
				mockChannelKeeperV2{packet: packet, found: tc.asyncFound},
			)

			err = mw.WriteAcknowledgement(ctx, destinationClient, sequence, tc.ack)
			if tc.expectedErr != "" {
				require.ErrorContains(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.expectWriteAck, writeAckWrapper.called)
			if tc.expectWriteAck {
				require.Equal(t, destinationClient, writeAckWrapper.client)
				require.Equal(t, sequence, writeAckWrapper.seq)
				require.Equal(t, tc.ack, writeAckWrapper.ack)
			}

			if tc.checkInflow {
				rateLimit, found := helper.App.RatelimitKeeper.GetRateLimit(ctx, rateLimitDenom, destinationClient)
				require.True(t, found)
				require.Equal(t, tc.expectedInflow, rateLimit.Flow.Inflow)

				found, err = helper.App.RatelimitKeeper.CheckPacketReceivedDuringCurrentQuota(ctx, destinationClient, sequence, rateLimitDenom)
				require.NoError(t, err)
				require.False(t, found)
			}
		})
	}
}
