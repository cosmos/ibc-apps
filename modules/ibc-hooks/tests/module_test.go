package tests

import (
	"fmt"
	"testing"

	_ "embed"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	ibc_hooks "github.com/cosmos/ibc-apps/modules/ibc-hooks/v7"
	"github.com/cosmos/ibc-apps/modules/ibc-hooks/v7/tests/helpers"
	"github.com/cosmos/ibc-apps/modules/ibc-hooks/v7/tests/mocks"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"
)

func TestOnRecvPacket(t *testing.T) {
	// create en env
	app, ctx, contractAddr, sender := helpers.SetupEnv(t)
	// Create the packet
	recvPacket := channeltypes.Packet{
		Data: transfertypes.FungibleTokenPacketData{
			Denom:    "transfer/channel-0/stake",
			Amount:   "1",
			Sender:   sender.GetAddress().String(),
			Receiver: contractAddr.String(),
			Memo:     fmt.Sprintf(`{"wasm":{"contract": "%s", "msg":{"increment":{}}}}`, contractAddr.String()),
		}.GetBytes(),
		SourcePort:    "transfer",
		SourceChannel: "channel-0",
	}
	// create the wasm hooks
	wasmHooks := ibc_hooks.NewWasmHooks(
		&app.IBCHooksKeeper,
		&app.WasmKeeper,
		"cosmos",
	)

	// create the ics4 middleware
	ics4Middleware := ibc_hooks.NewICS4Middleware(
		app.IBCKeeper.ChannelKeeper,
		wasmHooks,
	)

	// create the ibc middleware
	ibcMiddleware := ibc_hooks.NewIBCMiddleware(
		mocks.IBCModuleMock{},
		&ics4Middleware,
	)

	// call the hook
	res := ibcMiddleware.OnRecvPacket(
		ctx,
		recvPacket,
		sdk.MustAccAddressFromBech32("cosmos1ynjgz8t6vw2wjpra68g3jd85z6y5fea0uh7tvg"),
	)
	// assert the request was successful
	require.True(t, res.Success())

	// query the smart contract to ensure the count was set to 1
	// because it started with state None (counter/src/contract.rs:94)
	count, err := app.WasmKeeper.QuerySmart(
		ctx,
		contractAddr,
		[]byte(fmt.Sprintf(`{"get_count":{"addr": "%s"}}`, sender.GetAddress().String())),
	)
	require.NoError(t, err)
	require.Equal(t, `{"count":1}`, string(count))
}

func TestOnAcknowledgementPacket(t *testing.T) {
	app, ctx, contractAddr, sender := helpers.SetupEnv(t)
	recvPacket := channeltypes.Packet{
		Data: transfertypes.FungibleTokenPacketData{
			Denom:    "transfer/channel-0/stake",
			Amount:   "1",
			Sender:   sender.GetAddress().String(),
			Receiver: contractAddr.String(),
			Memo:     fmt.Sprintf(`{"wasm":{"contract": "%s", "msg":{"increment":{}}}}`, contractAddr),
		}.GetBytes(),
		SourcePort:    "transfer",
		SourceChannel: "channel-0",
	}

	// create the wasm hooks
	wasmHooks := ibc_hooks.NewWasmHooks(
		&app.IBCHooksKeeper,
		&app.WasmKeeper,
		"cosmos",
	)

	// create the ics4 middleware
	ics4Middleware := ibc_hooks.NewICS4Middleware(
		&mocks.ICS4WrapperMock{},
		wasmHooks,
	)

	// create the ibc middleware
	ibcMiddleware := ibc_hooks.NewIBCMiddleware(
		mocks.IBCModuleMock{},
		&ics4Middleware,
	)

	// call the hook
	seq, err := ibcMiddleware.SendPacket(
		ctx,
		&capabilitytypes.Capability{Index: 1},
		recvPacket.SourcePort,
		recvPacket.SourceChannel,
		ibcclienttypes.Height{
			RevisionNumber: 1,
			RevisionHeight: 1,
		},
		1,
		recvPacket.Data,
	)

	// require to be the first sequence
	require.Equal(t, uint64(1), seq)
	// assert the request was successful
	require.NoError(t, err)

	err = ibcMiddleware.WriteAcknowledgement(
		ctx,
		&capabilitytypes.Capability{Index: 1},
		nil,
		mocks.AcknowledgementMock{},
	)
	// assert the request was successful
	require.NoError(t, err)

	// query the smart contract to ensure the count was set to 0
	// because it started with state None (counter/src/contract.rs:94)
	count, err := app.WasmKeeper.QuerySmart(
		ctx,
		contractAddr,
		[]byte(fmt.Sprintf(`{"get_count":{"addr": %q}}`, sender.GetAddress().String())),
	)
	require.NoError(t, err)
	require.Equal(t, `{"count":10}`, string(count))
}
