package packetforward_test

import (
	"testing"

	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/packetforward"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/packetforward/keeper"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/test/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"
	porttypes "github.com/cosmos/ibc-go/v11/modules/core/05-port/types"
)

func TestUnmarshalPacketData(t *testing.T) {
	expPacketData := "packet data"
	app := mock.NewMockPacketUnmarshalerModule(gomock.NewController(t))
	app.EXPECT().
		UnmarshalPacketData(gomock.Any(), "transfer", "channel-0", []byte("data")).
		Return(expPacketData, transfertypes.V1, nil)

	middleware := packetforward.NewIBCMiddleware(app, &keeper.Keeper{}, 0, 0)

	packetData, version, err := middleware.UnmarshalPacketData(sdk.Context{}, "transfer", "channel-0", []byte("data"))
	require.NoError(t, err)
	require.Equal(t, transfertypes.V1, version)
	require.Equal(t, expPacketData, packetData)
}

func TestUnmarshalPacketData_AppDoesNotImplementUnmarshaler(t *testing.T) {
	middleware := packetforward.NewIBCMiddleware(struct{ porttypes.IBCModule }{}, &keeper.Keeper{}, 0, 0)

	_, _, err := middleware.UnmarshalPacketData(sdk.Context{}, "transfer", "channel-0", nil)
	require.ErrorIs(t, err, packetforward.ErrPacketDataUnmarshaler)
}
