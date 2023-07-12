package tests_unit

import (
	"encoding/json"
	"fmt"
	"testing"

	_ "embed"

	hooktypes "github.com/cosmos/ibc-apps/modules/ibc-hooks/v7/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	ibc_hooks "github.com/cosmos/ibc-apps/modules/ibc-hooks/v7"
	"github.com/cosmos/ibc-apps/modules/ibc-hooks/v7/simapp"
	"github.com/cosmos/ibc-apps/modules/ibc-hooks/v7/tests/unit/mocks"
	ibctransfer "github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibcmock "github.com/cosmos/ibc-go/v7/testing/mock"
	"github.com/stretchr/testify/suite"
)

//go:embed testdata/counter/artifacts/counter.wasm
var counterWasm []byte

//go:embed testdata/echo/artifacts/echo.wasm
var echoWasm []byte

//go:embed testdata/sender/artifacts/sender.wasm
var senderWasm []byte

type HooksTestSuite struct {
	suite.Suite

	App                 *simapp.App
	Ctx                 sdk.Context
	EchoContractAddr    sdk.AccAddress
	CounterContractAddr sdk.AccAddress
	SenderContractAddr  sdk.AccAddress
	TestAddress         *types.BaseAccount
}

func TestIBCHooksTestSuite(t *testing.T) {
	suite.Run(t, new(HooksTestSuite))
}

func (suite *HooksTestSuite) SetupEnv() {
	// Setup the environment
	app, ctx, acc := simapp.Setup(suite.T())

	// create the echo contract
	contractID, _, err := app.ContractKeeper.Create(ctx, acc.GetAddress(), counterWasm, nil)
	suite.NoError(err)
	counterContractAddr, _, err := app.ContractKeeper.Instantiate(
		ctx,
		contractID,
		acc.GetAddress(),
		nil,
		[]byte(`{"count": 0}`),
		"counter contract",
		nil,
	)
	suite.NoError(err)
	suite.Equal("cosmos14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s4hmalr", counterContractAddr.String())

	// create the counter contract
	contractID, _, err = app.ContractKeeper.Create(ctx, acc.GetAddress(), echoWasm, nil)
	suite.NoError(err)
	echoContractAddr, _, err := app.ContractKeeper.Instantiate(
		ctx,
		contractID,
		acc.GetAddress(),
		nil,
		[]byte(`{}`),
		"echo contract",
		nil,
	)
	suite.NoError(err)
	suite.Equal("cosmos1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrqez7la9", echoContractAddr.String())

	// create the counter contract
	contractID, _, err = app.ContractKeeper.Create(ctx, acc.GetAddress(), senderWasm, nil)
	suite.NoError(err)
	senderContractAddr, _, err := app.ContractKeeper.Instantiate(
		ctx,
		contractID,
		acc.GetAddress(),
		nil,
		[]byte(`{}`),
		"sender contract",
		nil,
	)
	suite.NoError(err)
	suite.Equal("cosmos17p9rzwnnfxcjp32un9ug7yhhzgtkhvl9jfksztgw5uh69wac2pgspeq65p", senderContractAddr.String())

	suite.App = app
	suite.Ctx = ctx
	suite.EchoContractAddr = echoContractAddr
	suite.CounterContractAddr = counterContractAddr
	suite.SenderContractAddr = senderContractAddr
	suite.TestAddress = acc
}

func (suite *HooksTestSuite) TestOnRecvPacketEcho() {
	// create en env
	suite.SetupEnv()

	// Create the packet
	recvPacket := channeltypes.Packet{
		Data: transfertypes.FungibleTokenPacketData{
			Denom:    "transfer/channel-0/stake",
			Amount:   "1",
			Sender:   suite.TestAddress.GetAddress().String(),
			Receiver: suite.EchoContractAddr.String(),
			Memo:     fmt.Sprintf(`{"wasm":{"contract": "%s", "msg":{"echo":{"msg":"test"}}}}`, suite.EchoContractAddr.String()),
		}.GetBytes(),
		SourcePort:    "transfer",
		SourceChannel: "channel-0",
	}

	// send funds to the escrow address to simulate a transfer from the ibc module
	escrowAddress := transfertypes.GetEscrowAddress(recvPacket.GetDestPort(), recvPacket.GetDestChannel())
	err := suite.App.BankKeeper.SendCoins(suite.Ctx, suite.TestAddress.GetAddress(), escrowAddress, sdk.NewCoins(sdk.NewInt64Coin("stake", 2)))
	suite.NoError(err)

	// create the wasm hooks
	wasmHooks := ibc_hooks.NewWasmHooks(
		&suite.App.IBCHooksKeeper,
		&suite.App.WasmKeeper,
		"cosmos",
	)

	// create the ics4 middleware
	ics4Middleware := ibc_hooks.NewICS4Middleware(
		suite.App.IBCKeeper.ChannelKeeper,
		wasmHooks,
	)

	// create the ibc middleware
	transferIBCModule := ibctransfer.NewIBCModule(suite.App.TransferKeeper)
	ibcmiddleware := ibc_hooks.NewIBCMiddleware(
		transferIBCModule,
		&ics4Middleware,
	)

	// call the hook twice
	res := ibcmiddleware.OnRecvPacket(
		suite.Ctx,
		recvPacket,
		suite.TestAddress.GetAddress(),
	)
	suite.True(res.Success())
	var ack map[string]string // This can't be unmarshalled to Acknowledgement because it's fetched from the events
	err = json.Unmarshal(res.Acknowledgement(), &ack)
	suite.Require().NoError(err)
	suite.Require().NotContains(ack, "error")
	suite.Require().Equal(ack["result"], "eyJjb250cmFjdF9yZXN1bHQiOiJkR2hwY3lCemFHOTFiR1FnWldOb2J3PT0iLCJpYmNfYWNrIjoiZXlKeVpYTjFiSFFpT2lKQlVUMDlJbjA9In0=")
}

func (suite *HooksTestSuite) TestOnRecvPacketCounterContract() {
	// create en env
	suite.SetupEnv()

	// Create the packet
	recvPacket := channeltypes.Packet{
		Data: transfertypes.FungibleTokenPacketData{
			Denom:    "transfer/channel-0/stake",
			Amount:   "1",
			Sender:   suite.TestAddress.GetAddress().String(),
			Receiver: suite.CounterContractAddr.String(),
			Memo:     fmt.Sprintf(`{"wasm":{"contract": "%s", "msg":{"increment":{}}}}`, suite.CounterContractAddr.String()),
		}.GetBytes(),
		SourcePort:    "transfer",
		SourceChannel: "channel-0",
	}

	// send funds to the escrow address to simulate a transfer from the ibc module
	escrowAddress := transfertypes.GetEscrowAddress(recvPacket.GetDestPort(), recvPacket.GetDestChannel())
	err := suite.App.BankKeeper.SendCoins(suite.Ctx, suite.TestAddress.GetAddress(), escrowAddress, sdk.NewCoins(sdk.NewInt64Coin("stake", 2)))
	suite.NoError(err)

	// create the wasm hooks
	wasmHooks := ibc_hooks.NewWasmHooks(
		&suite.App.IBCHooksKeeper,
		&suite.App.WasmKeeper,
		"cosmos",
	)

	// create the ics4 middleware
	ics4Middleware := ibc_hooks.NewICS4Middleware(
		suite.App.IBCKeeper.ChannelKeeper,
		wasmHooks,
	)

	// create the ibc middleware
	transferIBCModule := ibctransfer.NewIBCModule(suite.App.TransferKeeper)
	ibcmiddleware := ibc_hooks.NewIBCMiddleware(
		transferIBCModule,
		&ics4Middleware,
	)

	// call the hook twice
	res := ibcmiddleware.OnRecvPacket(
		suite.Ctx,
		recvPacket,
		suite.TestAddress.GetAddress(),
	)
	suite.True(res.Success())
	res = ibcmiddleware.OnRecvPacket(
		suite.Ctx,
		recvPacket,
		suite.TestAddress.GetAddress(),
	)
	suite.True(res.Success())
	// get the derived account to check the count
	senderBech32, err := suite.App.IBCHooksKeeper.DeriveIntermediateSender(
		suite.Ctx,
		recvPacket.GetDestChannel(),
		suite.TestAddress.GetAddress().String(),
		"cosmos",
		map[string]any{},
	)
	suite.NoError(err)
	// query the smart contract to assert the count
	count, err := suite.App.WasmKeeper.QuerySmart(
		suite.Ctx,
		suite.CounterContractAddr,
		[]byte(fmt.Sprintf(`{"get_count":{"addr": "%s"}}`, senderBech32)),
	)
	suite.NoError(err)
	suite.Equal(`{"count":1}`, string(count))
}

func (suite *HooksTestSuite) TestOnAcknowledgementPacketCounterContract() {
	suite.SetupEnv()

	callbackPacket := channeltypes.Packet{
		Data: transfertypes.FungibleTokenPacketData{
			Denom:    "transfer/channel-0/stake",
			Amount:   "1",
			Sender:   suite.TestAddress.GetAddress().String(),
			Receiver: suite.CounterContractAddr.String(),
			Memo:     fmt.Sprintf(`{"ibc_callback": "%s"}`, suite.CounterContractAddr),
		}.GetBytes(),
		Sequence:      1,
		SourcePort:    "transfer",
		SourceChannel: "channel-0",
	}

	// send funds to the escrow address to simulate a transfer from the ibc module
	escrowAddress := transfertypes.GetEscrowAddress(callbackPacket.GetDestPort(), callbackPacket.GetDestChannel())
	err := suite.App.BankKeeper.SendCoins(suite.Ctx, suite.TestAddress.GetAddress(), escrowAddress, sdk.NewCoins(sdk.NewInt64Coin("stake", 2)))
	suite.NoError(err)

	// create the wasm hooks
	wasmHooks := ibc_hooks.NewWasmHooks(
		&suite.App.IBCHooksKeeper,
		&suite.App.WasmKeeper,
		"cosmos",
	)

	// create the ics4 middleware
	ics4Middleware := ibc_hooks.NewICS4Middleware(
		&mocks.ICS4WrapperMock{},
		wasmHooks,
	)

	// create the ibc middleware
	transferIBCModule := ibctransfer.NewIBCModule(suite.App.TransferKeeper)
	ibcmiddleware := ibc_hooks.NewIBCMiddleware(
		transferIBCModule,
		&ics4Middleware,
	)

	// call the hook
	seq, err := ibcmiddleware.SendPacket(
		suite.Ctx,
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
	suite.Equal(uint64(1), seq)
	// assert the request was successful
	suite.NoError(err)

	// Create the packet
	recvPacket := channeltypes.Packet{
		Data: transfertypes.FungibleTokenPacketData{
			Denom:    "transfer/channel-0/stake",
			Amount:   "1",
			Sender:   suite.TestAddress.GetAddress().String(),
			Receiver: suite.CounterContractAddr.String(),
			Memo:     fmt.Sprintf(`{"wasm":{"contract": "%s", "msg":{"increment":{}}}}`, suite.CounterContractAddr.String()),
		}.GetBytes(),
		Sequence:      1,
		SourcePort:    "transfer",
		SourceChannel: "channel-0",
	}
	suite.NoError(err)
	err = wasmHooks.OnAcknowledgementPacketOverride(
		ibcmiddleware,
		suite.Ctx,
		recvPacket,
		ibcmock.MockAcknowledgement.Acknowledgement(),
		suite.TestAddress.GetAddress(),
	)
	// assert the request was successful
	suite.NoError(err)

	// query the smart contract to assert the count
	count, err := suite.App.WasmKeeper.QuerySmart(
		suite.Ctx,
		suite.CounterContractAddr,
		[]byte(fmt.Sprintf(`{"get_count":{"addr": %q}}`, suite.CounterContractAddr.String())),
	)
	suite.NoError(err)
	suite.Equal(`{"count":1}`, string(count))
}

func (suite *HooksTestSuite) TestOnTimeoutPacketOverrideCounterContract() {
	suite.SetupEnv()
	callbackPacket := channeltypes.Packet{
		Data: transfertypes.FungibleTokenPacketData{
			Denom:    "transfer/channel-0/stake",
			Amount:   "1",
			Sender:   suite.TestAddress.GetAddress().String(),
			Receiver: suite.CounterContractAddr.String(),
			Memo:     fmt.Sprintf(`{"ibc_callback": "%s"}`, suite.CounterContractAddr),
		}.GetBytes(),
		Sequence:      1,
		SourcePort:    "transfer",
		SourceChannel: "channel-0",
	}

	// send funds to the escrow address to simulate a transfer from the ibc module
	escrowAddress := transfertypes.GetEscrowAddress(callbackPacket.GetDestPort(), callbackPacket.GetDestChannel())
	err := suite.App.BankKeeper.SendCoins(suite.Ctx, suite.TestAddress.GetAddress(), escrowAddress, sdk.NewCoins(sdk.NewInt64Coin("stake", 2)))
	suite.NoError(err)

	// create the wasm hooks
	wasmHooks := ibc_hooks.NewWasmHooks(
		&suite.App.IBCHooksKeeper,
		&suite.App.WasmKeeper,
		"cosmos",
	)

	// create the ics4 middleware
	ics4Middleware := ibc_hooks.NewICS4Middleware(
		&mocks.ICS4WrapperMock{},
		wasmHooks,
	)

	// create the ibc middleware
	transferIBCModule := ibctransfer.NewIBCModule(suite.App.TransferKeeper)
	ibcmiddleware := ibc_hooks.NewIBCMiddleware(
		transferIBCModule,
		&ics4Middleware,
	)

	// call the hook
	seq, err := ibcmiddleware.SendPacket(
		suite.Ctx,
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
	suite.Equal(uint64(1), seq)
	// assert the request was successful
	suite.NoError(err)

	// Create the packet
	recvPacket := channeltypes.Packet{
		Data: transfertypes.FungibleTokenPacketData{
			Denom:    "transfer/channel-0/stake",
			Amount:   "1",
			Sender:   suite.TestAddress.GetAddress().String(),
			Receiver: suite.CounterContractAddr.String(),
			Memo:     fmt.Sprintf(`{"wasm":{"contract": "%s", "msg":{"increment":{}}}}`, suite.CounterContractAddr.String()),
		}.GetBytes(),
		Sequence:      1,
		SourcePort:    "transfer",
		SourceChannel: "channel-0",
	}
	suite.NoError(err)
	err = wasmHooks.OnTimeoutPacketOverride(
		ibcmiddleware,
		suite.Ctx,
		recvPacket,
		suite.TestAddress.GetAddress(),
	)
	// assert the request was successful
	suite.NoError(err)

	// query the smart contract to assert the count
	count, err := suite.App.WasmKeeper.QuerySmart(
		suite.Ctx,
		suite.CounterContractAddr,
		[]byte(fmt.Sprintf(`{"get_count":{"addr": %q}}`, suite.CounterContractAddr.String())),
	)
	suite.NoError(err)
	suite.Equal(`{"count":10}`, string(count))
}

func (suite *HooksTestSuite) TestOnRecvPacketAxelar() {
	// create en env
	suite.SetupEnv()

	sender := suite.TestAddress.GetAddress().String()

	suite.App.IBCHooksKeeper.SetParams(suite.Ctx, hooktypes.Params{
		Axelar: &hooktypes.Axelar{
			GmpAccount: sender,
			ChannelId:  "channel-10",
		},
	})

	// Create the packet
	recvPacket := channeltypes.Packet{
		Data: transfertypes.FungibleTokenPacketData{
			Denom:    "transfer/channel-0/stake",
			Amount:   "1",
			Sender:   sender,
			Receiver: suite.SenderContractAddr.String(),
			Memo: fmt.Sprintf(`{"wasm":{"source_address": "%s", "contract": "%s", "msg":{"sender":{}}}}`,
				"0x4673edfac441b98a9ad53c719e0309c09953910b",
				suite.SenderContractAddr.String(),
			),
		}.GetBytes(),
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		DestinationChannel: "channel-10",
	}

	// send funds to the escrow address to simulate a transfer from the ibc module
	escrowAddress := transfertypes.GetEscrowAddress(recvPacket.GetDestPort(), recvPacket.GetDestChannel())
	err := suite.App.BankKeeper.SendCoins(suite.Ctx, suite.TestAddress.GetAddress(), escrowAddress, sdk.NewCoins(sdk.NewInt64Coin("stake", 2)))
	suite.NoError(err)

	// create the wasm hooks
	wasmHooks := ibc_hooks.NewWasmHooks(
		&suite.App.IBCHooksKeeper,
		&suite.App.WasmKeeper,
		"cosmos",
	)

	// create the ics4 middleware
	ics4Middleware := ibc_hooks.NewICS4Middleware(
		suite.App.IBCKeeper.ChannelKeeper,
		wasmHooks,
	)

	// create the ibc middleware
	transferIBCModule := ibctransfer.NewIBCModule(suite.App.TransferKeeper)
	ibcmiddleware := ibc_hooks.NewIBCMiddleware(
		transferIBCModule,
		&ics4Middleware,
	)

	// call the hook twice
	res := ibcmiddleware.OnRecvPacket(
		suite.Ctx,
		recvPacket,
		suite.TestAddress.GetAddress(),
	)
	suite.True(res.Success())
	var ack map[string]string // This can't be unmarshalled to Acknowledgement because it's fetched from the events
	err = json.Unmarshal(res.Acknowledgement(), &ack)
	suite.Require().NoError(err)
	suite.Require().NotContains(ack, "error")
	// 0x4673edfac441b98a9ad53c719e0309c09953910b
	// cosmos1gee7m7kygxuc4xk483ceuqcfczv48ygtmkysru
	// {\"contract_result\":\"Y29zbW9zMWdlZTdtN2t5Z3h1YzR4azQ4M2NldXFjZmN6djQ4eWd0bWt5c3J1\",\"ibc_ack\":\"eyJyZXN1bHQiOiJBUT09In0=\"}
	suite.Require().Equal(ack["result"], "eyJjb250cmFjdF9yZXN1bHQiOiJZMjl6Ylc5ek1XZGxaVGR0TjJ0NVozaDFZelI0YXpRNE0yTmxkWEZqWm1ONmRqUTRlV2QwYld0NWMzSjEiLCJpYmNfYWNrIjoiZXlKeVpYTjFiSFFpT2lKQlVUMDlJbjA9In0=")
}
