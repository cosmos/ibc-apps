package v2

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10/keeper"
	"github.com/stretchr/testify/require"

	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypesv2 "github.com/cosmos/ibc-go/v10/modules/core/04-channel/v2/types"
	"github.com/cosmos/ibc-go/v10/modules/core/api"
)

type mockPacketUnmarshalerModule struct {
	api.IBCModule

	called     bool
	payload    channeltypesv2.Payload
	packetData any
}

func (m *mockPacketUnmarshalerModule) UnmarshalPacketData(payload channeltypesv2.Payload) (any, error) {
	m.called = true
	m.payload = payload
	return m.packetData, nil
}

func TestUnmarshalPacketData(t *testing.T) {
	payload := channeltypesv2.Payload{Value: []byte("payload")}
	expPacketData := "packet data"
	app := &mockPacketUnmarshalerModule{packetData: expPacketData}

	middleware := NewIBCMiddleware(keeper.Keeper{}, app)
	require.Implements(t, (*api.PacketDataUnmarshaler)(nil), middleware)

	packetData, err := middleware.UnmarshalPacketData(payload)
	require.NoError(t, err)
	require.True(t, app.called)
	require.Equal(t, payload, app.payload)
	require.Equal(t, expPacketData, packetData)
}

func TestUnmarshalPacketData_AppDoesNotImplementUnmarshaler(t *testing.T) {
	middleware := NewIBCMiddleware(keeper.Keeper{}, struct{ api.IBCModule }{})

	_, err := middleware.UnmarshalPacketData(channeltypesv2.Payload{})
	require.ErrorContains(t, err, "underlying application does not implement packet data unmarshaler")
}

func TestV2ToV1Packet_WithJSONEncoding(t *testing.T) {
	payloadValue := transfertypes.FungibleTokenPacketData{
		Denom:    "denom",
		Amount:   "100",
		Sender:   "sender",
		Receiver: "receiver",
		Memo:     "memo",
	}
	payloadValueBz, err := transfertypes.MarshalPacketData(payloadValue, transfertypes.V1, transfertypes.EncodingJSON)
	require.NoError(t, err)

	payload := channeltypesv2.Payload{
		SourcePort:      "sourcePort",
		DestinationPort: "destinationPort",
		Version:         transfertypes.V1,
		Encoding:        transfertypes.EncodingJSON,
		Value:           payloadValueBz,
	}

	v1Packet, err := v2ToV1Packet(payload, "sourceClient", "destinationClient", 1)
	require.NoError(t, err)
	require.Equal(t, uint64(1), v1Packet.Sequence)
	require.Equal(t, payload.SourcePort, v1Packet.SourcePort)
	require.Equal(t, "sourceClient", v1Packet.SourceChannel)
	require.Equal(t, payload.DestinationPort, v1Packet.DestinationPort)
	require.Equal(t, "destinationClient", v1Packet.DestinationChannel)

	var v1PacketData transfertypes.FungibleTokenPacketData
	err = json.Unmarshal(v1Packet.Data, &v1PacketData)
	require.NoError(t, err)
	require.Equal(t, payloadValue, v1PacketData)
}

func TestV2ToV1Packet_WithABIEncoding(t *testing.T) {
	payloadValue := transfertypes.FungibleTokenPacketData{
		Denom:    "denom",
		Amount:   "100",
		Sender:   "sender",
		Receiver: "receiver",
		Memo:     "memo",
	}

	payloadValueBz, err := transfertypes.MarshalPacketData(payloadValue, transfertypes.V1, transfertypes.EncodingABI)
	require.NoError(t, err)

	payload := channeltypesv2.Payload{
		SourcePort:      "sourcePort",
		DestinationPort: "destinationPort",
		Version:         transfertypes.V1,
		Encoding:        transfertypes.EncodingABI,
		Value:           payloadValueBz,
	}

	v1Packet, err := v2ToV1Packet(payload, "sourceClient", "destinationClient", 1)
	require.NoError(t, err)
	require.Equal(t, uint64(1), v1Packet.Sequence)
	require.Equal(t, payload.SourcePort, v1Packet.SourcePort)
	require.Equal(t, "sourceClient", v1Packet.SourceChannel)
	require.Equal(t, payload.DestinationPort, v1Packet.DestinationPort)
	require.Equal(t, "destinationClient", v1Packet.DestinationChannel)

	var v1PacketData transfertypes.FungibleTokenPacketData
	err = json.Unmarshal(v1Packet.Data, &v1PacketData)
	require.NoError(t, err)
	require.Equal(t, payloadValue, v1PacketData)
}

func TestV2ToV1Packet_WithProtobufEncoding(t *testing.T) {
	payloadValue := transfertypes.FungibleTokenPacketData{
		Denom:    "denom",
		Amount:   "100",
		Sender:   "sender",
		Receiver: "receiver",
		Memo:     "memo",
	}

	payloadValueBz, err := transfertypes.MarshalPacketData(payloadValue, transfertypes.V1, transfertypes.EncodingProtobuf)
	require.NoError(t, err)

	payload := channeltypesv2.Payload{
		SourcePort:      "sourcePort",
		DestinationPort: "destinationPort",
		Version:         transfertypes.V1,
		Encoding:        transfertypes.EncodingProtobuf,
		Value:           payloadValueBz,
	}

	v1Packet, err := v2ToV1Packet(payload, "sourceClient", "destinationClient", 1)
	require.NoError(t, err)
	require.Equal(t, uint64(1), v1Packet.Sequence)
	require.Equal(t, payload.SourcePort, v1Packet.SourcePort)
	require.Equal(t, "sourceClient", v1Packet.SourceChannel)
	require.Equal(t, payload.DestinationPort, v1Packet.DestinationPort)
	require.Equal(t, "destinationClient", v1Packet.DestinationChannel)

	var v1PacketData transfertypes.FungibleTokenPacketData
	err = json.Unmarshal(v1Packet.Data, &v1PacketData)
	require.NoError(t, err)
	require.Equal(t, payloadValue, v1PacketData)
}

func TestV2ToV1Packet_WithNilPayload(t *testing.T) {
	payload := channeltypesv2.Payload{
		SourcePort:      "sourcePort",
		DestinationPort: "destinationPort",
		Version:         transfertypes.V1,
		Encoding:        transfertypes.EncodingABI,
		Value:           nil,
	}

	_, err := v2ToV1Packet(payload, "sourceClient", "destinationClient", 1)
	require.Error(t, err)
}

func TestV2ToV1Packet_WithEmptyPayload(t *testing.T) {
	payload := channeltypesv2.Payload{
		SourcePort:      "sourcePort",
		DestinationPort: "destinationPort",
		Version:         transfertypes.V1,
		Encoding:        transfertypes.EncodingABI,
		Value:           []byte{},
	}

	_, err := v2ToV1Packet(payload, "sourceClient", "destinationClient", 1)
	require.Error(t, err)
}
