package ibc_hooks_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibchooks "github.com/cosmos/ibc-apps/modules/ibc-hooks/v11"
	"github.com/cosmos/ibc-apps/modules/ibc-hooks/v11/types"
	porttypes "github.com/cosmos/ibc-go/v11/modules/core/05-port/types"
	ibcmock "github.com/cosmos/ibc-go/v11/testing/mock"
)

func TestUnmarshalPacketData(t *testing.T) {
	ics4 := ibchooks.NewICS4Middleware(nil, nil)
	middleware := ibchooks.NewIBCMiddleware(&ibcmock.IBCModule{}, &ics4)

	packetData, version, err := middleware.UnmarshalPacketData(sdk.Context{}, "transfer", "channel-0", ibcmock.MockPacketData)
	require.NoError(t, err)
	require.Equal(t, ibcmock.Version, version)
	require.Equal(t, ibcmock.MockPacketData, packetData)
}

func TestUnmarshalPacketData_AppDoesNotImplementUnmarshaler(t *testing.T) {
	ics4 := ibchooks.NewICS4Middleware(nil, nil)
	middleware := ibchooks.NewIBCMiddleware(struct{ porttypes.IBCModule }{}, &ics4)

	_, _, err := middleware.UnmarshalPacketData(sdk.Context{}, "transfer", "channel-0", nil)
	require.ErrorIs(t, err, types.ErrPacketDataUnmarshaler)
}
