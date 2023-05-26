package tests_unit

import (
	"fmt"
	"testing"

	_ "embed"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	ibc_hooks "github.com/cosmos/ibc-apps/modules/ibc-hooks/v7"
	ibchookskeeper "github.com/cosmos/ibc-apps/modules/ibc-hooks/v7/keeper"
	"github.com/cosmos/ibc-apps/modules/ibc-hooks/v7/tests/unit/helpers"
	"github.com/cosmos/ibc-apps/modules/ibc-hooks/v7/tests/unit/mocks"
	ibctransfer "github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibcmock "github.com/cosmos/ibc-go/v7/testing/mock"
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

	// send funds to the escrow address to simulate a transfer from the ibc module
	escrowAddress := transfertypes.GetEscrowAddress(recvPacket.GetDestPort(), recvPacket.GetDestChannel())
	err := app.BankKeeper.SendCoins(ctx, sender.GetAddress(), escrowAddress, sdk.NewCoins(sdk.NewInt64Coin("stake", 2)))
	require.NoError(t, err)

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
	transferIBCModule := ibctransfer.NewIBCModule(app.TransferKeeper)
	ibcmiddleware := ibc_hooks.NewIBCMiddleware(
		transferIBCModule,
		&ics4Middleware,
	)

	// call the hook twice
	res := ibcmiddleware.OnRecvPacket(
		ctx,
		recvPacket,
		sender.GetAddress(),
	)
	require.True(t, res.Success())
	res = ibcmiddleware.OnRecvPacket(
		ctx,
		recvPacket,
		sender.GetAddress(),
	)
	require.True(t, res.Success())

	// get the derived account to check the count
	senderBech32, err := ibchookskeeper.DeriveIntermediateSender(
		recvPacket.GetDestChannel(),
		sender.GetAddress().String(),
		"cosmos",
	)
	require.NoError(t, err)
	// query the smart contract to assert the count
	count, err := app.WasmKeeper.QuerySmart(
		ctx,
		contractAddr,
		[]byte(fmt.Sprintf(`{"get_count":{"addr": "%s"}}`, senderBech32)),
	)
	require.NoError(t, err)
	require.Equal(t, `{"count":1}`, string(count))
}

func TestOnAcknowledgementPacket(t *testing.T) {
	app, ctx, contractAddr, sender := helpers.SetupEnv(t)
	callbackPacket := channeltypes.Packet{
		Data: transfertypes.FungibleTokenPacketData{
			Denom:    "transfer/channel-0/stake",
			Amount:   "1",
			Sender:   sender.GetAddress().String(),
			Receiver: contractAddr.String(),
			Memo:     fmt.Sprintf(`{"ibc_callback": "%s"}`, contractAddr),
		}.GetBytes(),
		Sequence:      1,
		SourcePort:    "transfer",
		SourceChannel: "channel-0",
	}

	// send funds to the escrow address to simulate a transfer from the ibc module
	escrowAddress := transfertypes.GetEscrowAddress(callbackPacket.GetDestPort(), callbackPacket.GetDestChannel())
	err := app.BankKeeper.SendCoins(ctx, sender.GetAddress(), escrowAddress, sdk.NewCoins(sdk.NewInt64Coin("stake", 2)))
	require.NoError(t, err)

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
	transferIBCModule := ibctransfer.NewIBCModule(app.TransferKeeper)
	ibcmiddleware := ibc_hooks.NewIBCMiddleware(
		transferIBCModule,
		&ics4Middleware,
	)

	// call the hook
	seq, err := ibcmiddleware.SendPacket(
		ctx,
		&capabilitytypes.Capability{Index: 1},
		callbackPacket.SourcePort,
		callbackPacket.SourceChannel,
		ibcclienttypes.Height{
			RevisionNumber: 1,
			RevisionHeight: 1,
		},
		1,
		callbackPacket.Data,
	)

	// require to be the first sequence
	require.Equal(t, uint64(1), seq)
	// assert the request was successful
	require.NoError(t, err)

	// Create the packet
	recvPacket := channeltypes.Packet{
		Data: transfertypes.FungibleTokenPacketData{
			Denom:    "transfer/channel-0/stake",
			Amount:   "1",
			Sender:   sender.GetAddress().String(),
			Receiver: contractAddr.String(),
			Memo:     fmt.Sprintf(`{"wasm":{"contract": "%s", "msg":{"increment":{}}}}`, contractAddr.String()),
		}.GetBytes(),
		Sequence:      1,
		SourcePort:    "transfer",
		SourceChannel: "channel-0",
	}
	require.NoError(t, err)
	err = wasmHooks.OnAcknowledgementPacketOverride(
		ibcmiddleware,
		ctx,
		recvPacket,
		ibcmock.MockAcknowledgement.Acknowledgement(),
		sender.GetAddress(),
	)
	// assert the request was successful
	require.NoError(t, err)

	// query the smart contract to assert the count
	count, err := app.WasmKeeper.QuerySmart(
		ctx,
		contractAddr,
		[]byte(fmt.Sprintf(`{"get_count":{"addr": %q}}`, contractAddr.String())),
	)
	require.NoError(t, err)
	require.Equal(t, `{"count":1}`, string(count))
}

func TestOnTimeoutPacketOverride(t *testing.T) {
	app, ctx, contractAddr, sender := helpers.SetupEnv(t)
	callbackPacket := channeltypes.Packet{
		Data: transfertypes.FungibleTokenPacketData{
			Denom:    "transfer/channel-0/stake",
			Amount:   "1",
			Sender:   sender.GetAddress().String(),
			Receiver: contractAddr.String(),
			Memo:     fmt.Sprintf(`{"ibc_callback": "%s"}`, contractAddr),
		}.GetBytes(),
		Sequence:      1,
		SourcePort:    "transfer",
		SourceChannel: "channel-0",
	}

	// send funds to the escrow address to simulate a transfer from the ibc module
	escrowAddress := transfertypes.GetEscrowAddress(callbackPacket.GetDestPort(), callbackPacket.GetDestChannel())
	err := app.BankKeeper.SendCoins(ctx, sender.GetAddress(), escrowAddress, sdk.NewCoins(sdk.NewInt64Coin("stake", 2)))
	require.NoError(t, err)

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
	transferIBCModule := ibctransfer.NewIBCModule(app.TransferKeeper)
	ibcmiddleware := ibc_hooks.NewIBCMiddleware(
		transferIBCModule,
		&ics4Middleware,
	)

	// call the hook
	seq, err := ibcmiddleware.SendPacket(
		ctx,
		&capabilitytypes.Capability{Index: 1},
		callbackPacket.SourcePort,
		callbackPacket.SourceChannel,
		ibcclienttypes.Height{
			RevisionNumber: 1,
			RevisionHeight: 1,
		},
		1,
		callbackPacket.Data,
	)

	// require to be the first sequence
	require.Equal(t, uint64(1), seq)
	// assert the request was successful
	require.NoError(t, err)

	// Create the packet
	recvPacket := channeltypes.Packet{
		Data: transfertypes.FungibleTokenPacketData{
			Denom:    "transfer/channel-0/stake",
			Amount:   "1",
			Sender:   sender.GetAddress().String(),
			Receiver: contractAddr.String(),
			Memo:     fmt.Sprintf(`{"wasm":{"contract": "%s", "msg":{"increment":{}}}}`, contractAddr.String()),
		}.GetBytes(),
		Sequence:      1,
		SourcePort:    "transfer",
		SourceChannel: "channel-0",
	}
	require.NoError(t, err)
	err = wasmHooks.OnTimeoutPacketOverride(
		ibcmiddleware,
		ctx,
		recvPacket,
		sender.GetAddress(),
	)
	// assert the request was successful
	require.NoError(t, err)

	// query the smart contract to assert the count
	count, err := app.WasmKeeper.QuerySmart(
		ctx,
		contractAddr,
		[]byte(fmt.Sprintf(`{"get_count":{"addr": %q}}`, contractAddr.String())),
	)
	require.NoError(t, err)
	require.Equal(t, `{"count":10}`, string(count))
}
