package simtests

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"

	wasmapp "github.com/CosmWasm/wasmd/app"
	"github.com/CosmWasm/wasmd/x/wasm/ibctesting"
	w "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	sdkibctesting "github.com/cosmos/ibc-go/v4/testing"
	"github.com/stretchr/testify/require"
)

type Chain struct {
	Chain  *ibctesting.TestChain
	Note   sdk.AccAddress
	Voice  sdk.AccAddress
	Tester sdk.AccAddress
}

type Suite struct {
	t *testing.T

	Coordinator *ibctesting.Coordinator
	ChainA      Chain
	ChainB      Chain
	ChainC      Chain
}

func SetupChain(t *testing.T, c *ibctesting.Coordinator, index int) Chain {
	chain := c.GetChain(sdkibctesting.GetChainID(index))
	chain.StoreCodeFile("../wasms/polytone_note.wasm")
	chain.StoreCodeFile("../wasms/polytone_voice.wasm")
	chain.StoreCodeFile("../wasms/polytone_proxy.wasm")
	chain.StoreCodeFile("../wasms/polytone_tester.wasm")

	// When the simulation environment executes a wasm message,
	// this is what is set as the max gas. For our testing
	// purposes, it doesn't matter that this isn't the real block
	// gas limit, only that we will be cut off upon reaching it,
	// as we're really trying to test that we save enough gas for
	// the reply.
	blockMaxGas := 2 * wasmapp.DefaultGas
	require.NotZero(t, blockMaxGas, "should be set")

	note := Instantiate(t, chain, 1, NoteInstantiate{
		BlockMaxGas: uint64(blockMaxGas),
	})
	voice := Instantiate(t, chain, 2, VoiceInstantiate{
		ProxyCodeId: 3,
		BlockMaxGas: uint64(blockMaxGas),
	})
	tester := Instantiate(t, chain, 4, TesterInstantiate{})

	return Chain{
		Chain:  chain,
		Note:   note,
		Voice:  voice,
		Tester: tester,
	}
}

func NewSuite(t *testing.T) Suite {
	coordinator := ibctesting.NewCoordinator(t, 3)
	chainA := SetupChain(t, coordinator, 0)
	chainB := SetupChain(t, coordinator, 1)
	chainC := SetupChain(t, coordinator, 2)

	return Suite{
		t:           t,
		Coordinator: coordinator,
		ChainA:      chainA,
		ChainB:      chainB,
		ChainC:      chainC,
	}
}

func ChannelConfig(port string) *sdkibctesting.ChannelConfig {
	return &sdkibctesting.ChannelConfig{
		PortID:  port,
		Version: "polytone-1",
		Order:   channeltypes.UNORDERED,
	}
}

func (c *Chain) QueryPort(contract sdk.AccAddress) string {
	return c.Chain.ContractInfo(contract).IBCPortID
}

func (s *Suite) SetupPath(aPort, bPort string, chainA, chainB *Chain) (*ibctesting.Path, error) {
	path := ibctesting.NewPath(chainA.Chain, chainB.Chain)
	path.EndpointA.ChannelConfig = ChannelConfig(aPort)
	path.EndpointB.ChannelConfig = ChannelConfig(bPort)

	// the ibctesting version of SetupPath does not return an
	// error, so we write it ourselves.
	setupClients := func(a, b *ibctesting.Endpoint) error {
		err := a.CreateClient()
		if err != nil {
			return err
		}
		err = b.CreateClient()
		if err != nil {
			return err
		}
		return nil

	}
	createConnections := func(a, b *ibctesting.Endpoint) error {
		err := a.ConnOpenInit()
		if err != nil {
			return err
		}
		err = b.ConnOpenTry()
		if err != nil {
			return err
		}
		err = a.ConnOpenAck()
		if err != nil {
			return err
		}
		err = b.ConnOpenConfirm()
		if err != nil {
			return err
		}
		err = a.UpdateClient()
		if err != nil {
			return err
		}
		return nil
	}
	createChannels := func(a, b *ibctesting.Endpoint) error {
		err := a.ChanOpenInit()
		if err != nil {
			return err
		}
		err = b.ChanOpenTry()
		if err != nil {
			return err
		}
		err = a.ChanOpenAck()
		if err != nil {
			return err
		}
		err = b.ChanOpenConfirm()
		if err != nil {
			return err
		}
		err = a.UpdateClient()
		if err != nil {
			return err
		}
		return nil
	}

	err := setupClients(path.EndpointA, path.EndpointB)
	if err != nil {
		return nil, err
	}
	err = createConnections(path.EndpointA, path.EndpointB)
	if err != nil {
		return nil, err
	}
	err = createChannels(path.EndpointA, path.EndpointB)
	if err != nil {
		return nil, err
	}
	return path, nil
}

func (s *Suite) SetupDefaultPath(
	chainA,
	chainB *Chain,
) *ibctesting.Path {
	// randomize the direction of the handshake. this should be a
	// no-op for a functional handshake.
	directionChoice := rand.Intn(2)

	aPort := chainA.QueryPort(chainA.Note)
	bPort := chainB.QueryPort(chainB.Voice)
	if directionChoice == 0 {
		// b -> a
		path, err := s.SetupPath(bPort, aPort, chainB, chainA)
		require.NoError(s.t, err)
		return path

	} else {
		// a -> b
		path, err := s.SetupPath(aPort, bPort, chainA, chainB)
		require.NoError(s.t, err)
		return path
	}
}

func (c *Chain) MintBondedDenom(t *testing.T, to sdk.AccAddress) {
	chain := c.Chain
	bondDenom := chain.App.StakingKeeper.BondDenom(chain.GetContext())
	coins := sdk.NewCoins(sdk.NewCoin(bondDenom, sdk.NewInt(100000000)))

	err := chain.App.BankKeeper.MintCoins(chain.GetContext(), minttypes.ModuleName, coins)
	require.NoError(t, err)

	err = chain.App.BankKeeper.SendCoinsFromModuleToAccount(chain.GetContext(), minttypes.ModuleName, to, coins)
	require.NoError(t, err)
}

func (s *Suite) RoundtripExecute(t *testing.T, path *ibctesting.Path, account *Account, msgs ...w.CosmosMsg) (CallbackDataExecute, error) {
	if msgs == nil {
		msgs = []w.CosmosMsg{}
	}
	msg := NoteExecuteMsg{
		Msgs:           msgs,
		TimeoutSeconds: 100,
		Callback: &CallbackRequest{
			Receiver: account.SuiteChain.Tester.String(),
			Msg:      "aGVsbG8K",
		},
	}
	callback, err := s.RoundtripMessage(t, path, account, NoteExecute{
		Execute: &msg,
	})
	if callback.FatalError != "" && err == nil {
		return callback.Execute, errors.New("internal error: " + callback.FatalError)
	}

	return callback.Execute, err
}

func (s *Suite) RoundtripQuery(t *testing.T, path *ibctesting.Path, account *Account, msgs ...w.QueryRequest) (CallbackDataQuery, error) {
	if msgs == nil {
		msgs = []w.QueryRequest{}
	}
	msg := NoteExecuteQuery{
		Msgs:           msgs,
		TimeoutSeconds: 100,
		Callback: CallbackRequest{
			Receiver: account.SuiteChain.Tester.String(),
			Msg:      "aGVsbG8K",
		},
	}
	callback, err := s.RoundtripMessage(t, path, account, NoteExecute{
		Query: &msg,
	})
	if callback.FatalError != "" && err == nil {
		return callback.Query, errors.New(callback.FatalError)
	}
	return callback.Query, err
}

func (s *Suite) RoundtripMessage(t *testing.T, path *ibctesting.Path, account *Account, msg NoteExecute) (Callback, error) {
	startCallbacks := QueryCallbackHistory(account.Chain, account.SuiteChain.Tester)
	wasmMsg := account.WasmExecute(&account.SuiteChain.Note, msg)
	if _, err := account.Send(t, wasmMsg); err != nil {
		return Callback{}, err
	}
	if err := s.Coordinator.RelayAndAckPendingPackets(path); err != nil {
		return Callback{}, err
	}
	callbacks := QueryCallbackHistory(account.Chain, account.SuiteChain.Tester)
	require.Equal(t, len(startCallbacks)+1, len(callbacks), "no new callbacks")
	callback := callbacks[len(callbacks)-1]
	require.Equal(t, account.Address.String(), callback.Initiator)

	return callback.Result, nil
}

func HelloMessage(to sdk.AccAddress, data string) w.CosmosMsg {
	return w.CosmosMsg{
		Wasm: &w.WasmMsg{
			Execute: &w.ExecuteMsg{
				ContractAddr: to.String(),
				Msg: []byte(
					fmt.Sprintf(`{"hello": { "data": "%s" }}`,
						data,
					)),
				Funds: []w.Coin{},
			},
		},
	}
}
