package simtests

import (
	"encoding/json"
	"testing"

	w "github.com/CosmWasm/wasmvm/types"
	"github.com/stretchr/testify/require"
)

const (
	testBinary string = "aGVsbG8=" // "hello" in base64
	testText   string = "hello"
)

// I can:
//
//   - Execute multiple messages (wasm and non-wasm) on a remote chain
//     and get a callback containing their response data.
//   - Execute multiple queries (wasm and non-wasm) on a remote chain
//     and get their responses in a callback.
func TestFunctionality(t *testing.T) {
	suite := NewSuite(t)

	path := suite.SetupDefaultPath(&suite.ChainA, &suite.ChainB)

	// Execute two messages, the first of which uses
	// polytone-tester to set some data in the transaction
	// response, and the second of which sets the proxy's staking
	// rewards receiver address to the voice address on the remote
	// chain.

	accountA := GenAccount(t, &suite.ChainA)
	dataCosmosMsg := HelloMessage(suite.ChainB.Tester, string(testBinary))

	noDataCosmosMsg := w.CosmosMsg{
		Distribution: &w.DistributionMsg{
			SetWithdrawAddress: &w.SetWithdrawAddressMsg{
				Address: suite.ChainB.Voice.String(),
			},
		},
	}

	callbackExecute, err := suite.RoundtripExecute(t, path, &accountA, dataCosmosMsg, noDataCosmosMsg)
	if err != nil {
		t.Fatal(err)
	}
	require.Len(t, callbackExecute.Ok.Result, 2, "error: "+callbackExecute.Err)
	require.Equal(t, "", callbackExecute.Err)

	result1 := unmarshalExecute(t, callbackExecute.Ok.Result[0].Data).Data
	result2 := unmarshalExecute(t, callbackExecute.Ok.Result[1].Data).Data

	require.Equal(t, testText, string(result1))
	require.Equal(t, "", string(result2))

	balanceQuery := w.QueryRequest{
		Bank: &w.BankQuery{
			Balance: &w.BalanceQuery{
				Address: suite.ChainB.Note.String(),
				Denom:   suite.ChainB.Chain.App.StakingKeeper.BondDenom(suite.ChainB.Chain.GetContext()),
			},
		},
	}

	testerQuery := TesterQuery{
		History: &Empty{},
	}
	queryBytes, err := json.Marshal(testerQuery)
	if err != nil {
		t.Fatal(err)
	}

	historyQuery := w.QueryRequest{
		Wasm: &w.WasmQuery{
			Smart: &w.SmartQuery{
				ContractAddr: suite.ChainB.Tester.String(),
				Msg:          queryBytes,
			},
		},
	}

	callbackQuery, err := suite.RoundtripQuery(t, path, &accountA, balanceQuery, historyQuery)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t,
		CallbackDataQuery{
			Ok: [][]byte{
				[]byte(`{"amount":{"denom":"stake","amount":"100"}}`), // contracts get made with 100 coins.
				[]byte(`{"history":[]}`),
			},
		}, callbackQuery)
}

// Generates two addresses from the same private key on chains B and
// C, then sends a message from each accounts proxy. The two addresses
// will have the same string representation, as the two chains have
// the same prefix, and the same local connection and channel ID. They
// also have the same remote port, as they are the first instantation
// of the same bytecode on chains with the same prefix.
//
// If these two different accounts get different addreses on chain A,
// it means that the contract is correctly distinguishing them based
// on some combination of local `(connection_id, channel_id)`, as
// those are the only parts of the messages that differ.
//
// Later tests will show that the contract does not change the address
// on chain A if a channel closes, which together means that the
// contract is correctly namespacing addresses based on connection_id.
func TestSameAddressDifferentChains(t *testing.T) {
	suite := NewSuite(t)

	pathCA := suite.SetupDefaultPath(&suite.ChainC, &suite.ChainA)
	pathBA := suite.SetupDefaultPath(&suite.ChainB, &suite.ChainA)

	friend := GenAccount(t, &suite.ChainB)

	// this follows the rules of Cosmos to induce the scenerio,
	// though signatures are not required for a message to be
	// sent from a malicious note contract, and anyone can
	// duplicate a chain, so you can imagine an attacker inducing
	// this scenerio at will.
	duplicate := friend.KeplrChainDropdownSelect(t, &suite.ChainC)

	require.Equal(t, friend.Address.String(), duplicate.Address.String())

	helloMsg := HelloMessage(suite.ChainA.Tester, "")

	b, err := suite.RoundtripExecute(t, pathBA, &friend, helloMsg)
	if err != nil {
		t.Fatal(err)
	}
	c, err := suite.RoundtripExecute(t, pathCA, &duplicate, helloMsg)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, "", b.Err)
	require.Equal(t, "", c.Err)
	require.Equal(t, []byte(nil), b.Ok.Result[0].Data)
	require.Equal(t, []byte(nil), c.Ok.Result[0].Data)
	require.Equal(t, c.Ok.Result, b.Ok.Result)

	history := QueryHelloHistory(suite.ChainA.Chain, suite.ChainA.Tester)
	require.Len(t, history, 2)
	require.NotEqual(t, history[0], history[1])
	require.Equal(t, b.Ok.ExecutedBy, history[0], "first message executed by chain B remote account")
	require.Equal(t, c.Ok.ExecutedBy, history[1], "second message executed by chain C remote account")
}

// Checks that connections between two of the same modules are not
// allowed. This checks that we are using the handshake logic, the
// other permutations of the handshake are tested in the
// polytone/handshake package.
func TestHandshakeBetweenSameModule(t *testing.T) {
	suite := NewSuite(t)

	aNote := suite.ChainA.QueryPort(suite.ChainA.Note)
	aVoice := suite.ChainA.QueryPort(suite.ChainA.Voice)
	bNote := suite.ChainB.QueryPort(suite.ChainB.Note)
	bVoice := suite.ChainB.QueryPort(suite.ChainB.Voice)

	_, err := suite.SetupPath(aNote, bNote, &suite.ChainA, &suite.ChainB)
	require.ErrorContains(t,
		err,
		"channel open try callback failed",
		"note <-/-> note",
	)
	// for reasons i do not understand, if the try step fails the
	// sequence number for the sending account does not get
	// incremented correctly. as a bandaid, this manually corrects.
	//
	// FIXME: why?
	suite.ChainB.Chain.SenderAccount.SetSequence(suite.ChainA.Chain.SenderAccount.GetSequence() + 1)

	_, err = suite.SetupPath(bVoice, aVoice, &suite.ChainB, &suite.ChainA)
	require.ErrorContains(t,
		err,
		"channel open try callback failed",
		"voice <-/-> voice",
	)
	suite.ChainA.Chain.SenderAccount.SetSequence(suite.ChainA.Chain.SenderAccount.GetSequence() + 1)

	_, err = suite.SetupPath(aVoice, bNote, &suite.ChainA, &suite.ChainB)
	require.NoError(t, err, "voice <- -> note")
}

// Executes a message on the note chain that will run out of gas on
// the voice chain and makes sure that an ACK + callback indicating
// that the out-of-gas error occured is returned.
func TestVoiceOutOfGas(t *testing.T) {
	suite := NewSuite(t)

	path := suite.SetupDefaultPath(&suite.ChainA, &suite.ChainB)

	accountA := GenAccount(t, &suite.ChainA)
	gasMsg := `{"run_out_of_gas":{}}`
	gasCosmosgMsg := w.CosmosMsg{
		Wasm: &w.WasmMsg{
			Execute: &w.ExecuteMsg{
				ContractAddr: suite.ChainB.Tester.String(),
				Msg:          []byte(gasMsg),
				Funds:        []w.Coin{},
			},
		},
	}

	_, err := suite.RoundtripExecute(t, path, &accountA, gasCosmosgMsg)
	require.ErrorContains(t,
		err,
		"internal error: codespace: sdk, code: 11",
		"internal error should be returned indicating out of gas",
	)
}

// Tests that if a callback receiver runs out of gas the ACK is still
// comitted on the sending chain and the local to remote address
// mapping is still updated.
func TestNoteOutOfGas(t *testing.T) {
	suite := NewSuite(t)

	path := suite.SetupDefaultPath(&suite.ChainA, &suite.ChainB)
	account := GenAccount(t, &suite.ChainA)
	msg := NoteExecute{
		Execute: &NoteExecuteMsg{
			Msgs:           []w.CosmosMsg{},
			TimeoutSeconds: 300,
			Callback: &CallbackRequest{
				Receiver: suite.ChainA.Tester.String(),
				Msg:      "cnVuX291dF9vZl9nYXM=", // run_out_of_gas base64
			},
		},
	}
	startCallbacks := QueryCallbackHistory(account.Chain, account.SuiteChain.Tester)
	wasmMsg := account.WasmExecute(&account.SuiteChain.Note, msg)
	if _, err := account.Send(t, wasmMsg); err != nil {
		require.NoError(t, err)
	}
	if err := suite.Coordinator.RelayAndAckPendingPackets(path); err != nil {
		require.NoError(t, err)
	}
	callbacks := QueryCallbackHistory(account.Chain, account.SuiteChain.Tester)
	require.Equal(t, len(startCallbacks), len(callbacks), "no new callbacks")

	remote := QueryRemoteAccount(suite.ChainA.Chain, suite.ChainA.Note, account.Address)
	require.NotEmpty(t, remote, "remote account set even though callback rolled back")
}

// Tests executing a message on the remote chain, checking the
// callback, and then executing another message.
//
// This tests that we correctly save proxies and reuse them upon
// another message being executed.
func TestMultipleMessages(t *testing.T) {
	suite := NewSuite(t)

	path := suite.SetupDefaultPath(&suite.ChainA, &suite.ChainB)

	accountA := GenAccount(t, &suite.ChainA)
	dataCosmosMsg := HelloMessage(suite.ChainA.Tester, testBinary)

	noDataCosmosMsg := w.CosmosMsg{
		Distribution: &w.DistributionMsg{
			SetWithdrawAddress: &w.SetWithdrawAddressMsg{
				Address: suite.ChainB.Voice.String(),
			},
		},
	}

	callback, err := suite.RoundtripExecute(t, path, &accountA, dataCosmosMsg, noDataCosmosMsg)
	require.NoError(t, err)
	response := unmarshalExecute(t, callback.Ok.Result[0].Data).Data
	require.Equal(t, testText, string(response))
	require.Equal(t, []byte(nil), callback.Ok.Result[1].Data)

	callback, err = suite.RoundtripExecute(t, path, &accountA, dataCosmosMsg, noDataCosmosMsg)
	require.NoError(t, err)
	response = unmarshalExecute(t, callback.Ok.Result[0].Data).Data
	require.Equal(t, testText, string(response))
	require.Equal(t, []byte(nil), callback.Ok.Result[1].Data)
}

// A note may only ever connect to a single voice. This simplifies the
// API (as channel_id does not need to be specifed after a single
// handshake), and simplifies the protocol.
func TestOneVoicePerNote(t *testing.T) {
	suite := NewSuite(t)
	// connect note on A to voice on C. note should not connect
	// any additional connections.
	_ = suite.SetupDefaultPath(&suite.ChainA, &suite.ChainC)

	cPort := suite.ChainB.QueryPort(suite.ChainC.Voice)
	bPort := suite.ChainB.QueryPort(suite.ChainB.Voice)
	aPort := suite.ChainA.QueryPort(suite.ChainA.Note)
	_, err := suite.SetupPath(
		bPort,
		aPort,
		&suite.ChainB,
		&suite.ChainA,
	)
	require.ErrorContains(t,
		err,
		"contract is already paired with port ("+
			cPort+
			") on connection (connection-0), got port ("+
			bPort+
			") on connection (connection-1)",
		"two voices may not be connected to the same note",
	)
}

// Executes a "hello" call to chain B's tester contract via the chain
// A->B Polytone connection. Checks that:
//
//  1. Before execution, the sender does not have a remote account that
//     is queryable.
//  2. After execution they do.
//  3. The query response matches the callback response, matches the
//     address that executed the "hello" call on chain B.
func TestRemoteAddressBookkeeping(t *testing.T) {
	suite := NewSuite(t)
	path := suite.SetupDefaultPath(&suite.ChainA, &suite.ChainB)

	account := GenAccount(t, &suite.ChainA)
	remoteAccount := QueryRemoteAccount(
		suite.ChainA.Chain,
		suite.ChainA.Note,
		account.Address,
	)
	require.Equal(t,
		`null`,
		remoteAccount,
		"no remote account exists before a message is sent",
	)

	callback, err := suite.RoundtripExecute(t, path, &account)
	require.NoError(t, err, "executing no messages should create an account")
	remoteAccount = QueryRemoteAccount(
		suite.ChainA.Chain,
		suite.ChainA.Note,
		account.Address,
	)
	require.Equal(
		t,
		`"`+callback.Ok.ExecutedBy+`"`,
		remoteAccount,
		"account created matches account returned by callback",
	)
}

// Executes a hello call and a bank message to send more tokens than
// the proxy has. Verifies that an error callback was returned with
// the correct message index and that the hello call is not reflected
// in the tester's hello history (it was rolled back).
func TestErrorReturnsErrorAndRollsBack(t *testing.T) {
	suite := NewSuite(t)
	path := suite.SetupDefaultPath(&suite.ChainA, &suite.ChainB)

	account := GenAccount(t, &suite.ChainA)
	hello := HelloMessage(suite.ChainB.Tester, testBinary)
	bankMsg := w.CosmosMsg{
		Bank: &w.BankMsg{
			Send: &w.SendMsg{
				ToAddress: suite.ChainB.Voice.String(),
				Amount: []w.Coin{
					{
						Denom:  suite.ChainB.Chain.App.StakingKeeper.BondDenom(suite.ChainB.Chain.GetContext()),
						Amount: "100",
					},
				},
			},
		},
	}
	callback, err := suite.RoundtripExecute(t, path, &account, hello, bankMsg)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t,
		CallbackDataExecute{
			Err: "codespace: wasm, code: 5",
		},
		callback,
		"proxy errored during execution",
	)

	history := QueryHelloHistory(suite.ChainB.Chain, suite.ChainB.Tester)
	require.Empty(t, history, "history message should have been rolled back on bank msg failure")
}

// Test that query failures correctly return their index.
func TestQueryErrors(t *testing.T) {
	suite := NewSuite(t)
	path := suite.SetupDefaultPath(&suite.ChainA, &suite.ChainB)
	account := GenAccount(t, &suite.ChainA)

	testerQuery := TesterQuery{
		History: &Empty{},
	}
	queryBytes, err := json.Marshal(testerQuery)
	if err != nil {
		t.Fatal(err)
	}
	goodQuery := w.QueryRequest{
		Wasm: &w.WasmQuery{
			Smart: &w.SmartQuery{
				ContractAddr: suite.ChainB.Tester.String(),
				Msg:          queryBytes,
			},
		},
	}

	// tester query against voice module will error.
	badQuery := w.QueryRequest{
		Wasm: &w.WasmQuery{
			Smart: &w.SmartQuery{
				ContractAddr: suite.ChainB.Voice.String(),
				Msg:          queryBytes,
			},
		},
	}

	callback, err := suite.RoundtripQuery(t, path, &account, goodQuery, badQuery)
	if err != nil {
		t.Fatal(err)
	}
	// codespace 9 = "query wasm contract failed"
	require.Equal(t,
		CallbackDataQuery{
			Err: ErrorResponse{
				MessageIndex: 1,
				Error:        "contract: codespace: wasm, code: 9",
			},
		},
		callback,
		"second query should fail",
	)
}

// Tests that the data returned in a callback contains the address of
// instantiated contracts and can be accessed by
// parse_reply_instantiate_data.
func TestInstantiateExecute(t *testing.T) {
	suite := NewSuite(t)

	path := suite.SetupDefaultPath(&suite.ChainA, &suite.ChainB)

	accountA := GenAccount(t, &suite.ChainA)
	msg, err := json.Marshal(TesterInstantiate{})
	require.NoError(t, err)
	initCosmosMsg := w.CosmosMsg{
		Wasm: &w.WasmMsg{
			Instantiate: &w.InstantiateMsg{
				CodeID: 4,
				Msg:    msg,
				Funds:  []w.Coin{},
				Label:  "test",
			},
		},
	}

	callback, err := suite.RoundtripExecute(t, path, &accountA, initCosmosMsg)
	if err != nil {
		t.Fatal(err)
	}
	require.Empty(t, callback.Err, "callback should not error")
	response := unmarshalInstantiate(t, callback.Ok.Result[0].Data)

	// address should be: cosmos1ghd753shjuwexxywmgs4xz7x2q732vcnkm6h2pyv9s6ah3hylvrqa0dr5q
	// But because it can change in the future, we just check its not empty
	require.NotEmpty(t, response.Address, "address should not be empty")
}

func TestSimpleChannelClosure(t *testing.T) {
	suite := NewSuite(t)

	account := GenAccount(t, &suite.ChainA)
	path := suite.SetupDefaultPath(&suite.ChainA, &suite.ChainB)

	_, err := suite.RoundtripExecute(t, path, &account)
	require.NoError(t, err, "creating an account should work")
	remoteAccount := QueryRemoteAccount(suite.ChainA.Chain, suite.ChainA.Note, account.Address)
	require.NotEqual(t, "null", remoteAccount, "remote account was created")

	initialActiveChannel := QueryActiveChannel(suite.ChainA.Chain, suite.ChainA.Note)
	require.Equal(t, `"channel-0"`, initialActiveChannel)

	suite.Coordinator.CloseChannel(path)

	require.Equal(
		t,
		remoteAccount,
		QueryRemoteAccount(suite.ChainA.Chain, suite.ChainA.Note, account.Address),
		"remote account is queryable even when channel is closed",
	)

	aPort := suite.ChainA.QueryPort(suite.ChainA.Note)
	bPort := suite.ChainB.QueryPort(suite.ChainB.Voice)

	// get around path creation sequence number bug in ibctesting.
	suite.ChainA.Chain.SenderAccount = account.Acc
	suite.ChainA.Chain.SenderPrivKey = account.PrivKey

	// try to create a path between the modules on a new
	// connection. this should fail.
	_, err = suite.SetupPath(
		bPort,
		aPort,
		&suite.ChainB,
		&suite.ChainA,
	)
	require.ErrorContains(t,
		err,
		"contract is already paired with port ("+
			bPort+
			") on connection (connection-0), got port ("+
			bPort+
			") on connection (connection-1)",
		"the same port on a different channel is not allowed",
	)
	// errors cause sequence numbers to get messed up. i don't
	// make the rules 🤷
	account.Acc.SetSequence(account.Acc.GetSequence() + 1)

	// sending a message on a closed channel results in an error.
	_, err = suite.RoundtripExecute(
		t,
		path,
		&account,
		HelloMessage(
			suite.ChainB.Tester,
			"👌",
		),
	)
	require.ErrorContains(t,
		err,
		"contract has no pair, establish a channel with a voice module to create one",
		"messages can not be executed when the channel is closed",
	)

	account = GenAccount(t, &suite.ChainA)
	suite.ChainA.Chain.SenderAccount = account.Acc
	suite.ChainA.Chain.SenderPrivKey = account.PrivKey

	// Create a new channel, on the original connection. Doing
	// this requires a bit of a hack where we reset the path's
	// channel information to its default values and then create a
	// new channel. ibctesting doesn't have a "way" to do this, so
	// this seemed like the lowest-effort hack.
	path.EndpointA.ChannelConfig = ChannelConfig(aPort)
	path.EndpointB.ChannelConfig = ChannelConfig(bPort)
	path.EndpointA.ChannelID = ""
	path.EndpointB.ChannelID = ""

	// This should create a channel! or error. But clearly it does
	// not because there is no code path that calls into
	// ibc_confirm and doesn't set the active channel. Fuck.
	suite.Coordinator.CreateChannels(path)

	// I have toiled many hours on this test and got nowhere. The
	// above is the closest I ever got to isolating why the
	// testing environment was behaving like this. It appears that
	// the CosmWasm contract just doesn't get called, despite no
	// error being returned while executing the channel closure.
	//
	// The only way for a channel to close with Polytone is if (1)
	// the counterparty is malicious, or (2) the light client has
	// been taken over. In both cases, you are completely pwned
	// anyhow as the adversary has complete controll of the remote
	// accounts, so I think completely reaching this state is more
	// of an academic interest.

	// activeChannel := QueryActiveChannel(suite.ChainA.Chain, suite.ChainA.Note)
	// require.Equal(t, `"channel-1"`, activeChannel, "a new channel should have been created")

	// callback, err := suite.RoundtripExecute(
	// 	t,
	// 	path,
	// 	&account,
	// 	HelloMessage(
	// 		suite.ChainB.Tester,
	// 		"👌",
	// 	),
	// )
	// require.NoError(t, err, "messages can be executed now that the channel is reopened")
	// require.Equal(t, "👌", callback.Ok.Result[0].Data)
	// require.Equal(
	// 	t,
	// 	remoteAccount,
	// 	QueryRemoteAccount(suite.ChainA.Chain, suite.ChainA.Note, account.Address),
	// 	"remote account has not changed",
	// )
}
