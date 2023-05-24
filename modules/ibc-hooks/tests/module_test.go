package tests

import (
	"testing"

	_ "embed"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibc_hooks "github.com/cosmos/ibc-apps/modules/ibc-hooks/v7"
	"github.com/cosmos/ibc-apps/modules/ibc-hooks/v7/simapp"
	"github.com/cosmos/ibc-apps/modules/ibc-hooks/v7/tests/mocks"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/counter/artifacts/counter.wasm
var counterWasm []byte

var Sender = sdk.AccAddress([]byte("cosmos1ynjgz8t6vw2wjpra68g3jd85z6y5fea0uh7tvg"))
var packet = channeltypes.Packet{
	Data: transfertypes.FungibleTokenPacketData{
		Denom:    "transfer/channel-0/uatom",
		Amount:   "0",
		Sender:   Sender.String(),
		Receiver: "cosmos14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s4hmalr",
		Memo:     `{"wasm":{"contract": "cosmos14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s4hmalr", "msg":{"increment":{}}}}`,
	}.GetBytes(),
	SourcePort:    "transfer",
	SourceChannel: "channel-0",
}

func TestOnRecvPacketOverride(t *testing.T) {
	// Setup the environment
	app := simapp.Setup(t)
	ctx, wasmcontractkeeper := wasmkeeper.CreateTestInput(t, false, "iterator,staking,stargate,cosmwasm_1_1")

	// create the contract because it cannot be mocked
	contractID, _, err := wasmcontractkeeper.ContractKeeper.Create(ctx, Sender, counterWasm, nil)
	require.NoError(t, err)
	addr, _, err := wasmcontractkeeper.ContractKeeper.Instantiate(
		ctx,
		contractID,
		Sender,
		nil,
		[]byte(`{"count": 0}`),
		"demo contract 3",
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, "cosmos14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s4hmalr", addr.String())
	// create the ibc middleware
	ibcMiddleware := ibc_hooks.NewIBCMiddleware(mocks.IBCModuleMock{}, nil)
	// create the wasm hooks
	wasmHooks := ibc_hooks.NewWasmHooks(
		&app.IBCHooksKeeper,
		wasmcontractkeeper.WasmKeeper,
		"cosmos",
	)
	// call the hook
	res := wasmHooks.OnRecvPacketOverride(
		ibcMiddleware,
		ctx,
		packet,
		sdk.MustAccAddressFromBech32("cosmos1ynjgz8t6vw2wjpra68g3jd85z6y5fea0uh7tvg"),
	)
	// assert the request was successful
	require.True(t, res.Success())
}
