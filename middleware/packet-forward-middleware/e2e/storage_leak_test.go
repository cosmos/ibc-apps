package e2e

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"cosmossdk.io/math"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	chantypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/relayer"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

type PFMExport struct {
	AppState struct {
		PacketForwardMiddleware struct {
			InFlightPackets map[string]interface{} `json:"in_flight_packets"`
		} `json:"packetfowardmiddleware"`
	} `json:"app_state"`
}

// TestStorageLeak verifies that that the PFM module does not retain any in-flight packets after a timeout occurs
func TestStorageLeak(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	var (
		ctx                          = context.Background()
		client, network              = interchaintest.DockerSetup(t)
		rep                          = testreporter.NewNopReporter()
		eRep                         = rep.RelayerExecReporter(t)
		chainIdA, chainIdB, chainIdC = "chain-a", "chain-b", "chain-c"
	)

	vals := 1
	fullNodes := 0

	baseCfg := DefaultConfig

	baseCfg.ChainID = chainIdA
	configA := baseCfg

	baseCfg.ChainID = chainIdB
	configB := baseCfg

	baseCfg.ChainID = chainIdC
	configC := baseCfg

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{Name: "pfm", ChainConfig: configA, NumFullNodes: &fullNodes, NumValidators: &vals},
		{Name: "pfm", ChainConfig: configB, NumFullNodes: &fullNodes, NumValidators: &vals},
		{Name: "pfm", ChainConfig: configC, NumFullNodes: &fullNodes, NumValidators: &vals},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	chainA, chainB, chainC := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain), chains[2].(*cosmos.CosmosChain)

	r := interchaintest.NewBuiltinRelayerFactory(
		ibc.Hermes,
		zaptest.NewLogger(t),
		relayer.DockerImage(&DefaultRelayer),
	).Build(t, client, network)

	const pathAB = "ab"
	const pathBC = "bc"

	ic := interchaintest.NewInterchain().
		AddChain(chainA).
		AddChain(chainB).
		AddChain(chainC).
		AddRelayer(r, "relayer").
		AddLink(interchaintest.InterchainLink{
			Chain1:  chainA,
			Chain2:  chainB,
			Relayer: r,
			Path:    pathAB,
		}).
		AddLink(interchaintest.InterchainLink{
			Chain1:  chainB,
			Chain2:  chainC,
			Relayer: r,
			Path:    pathBC,
		})

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: false,
	}))

	t.Cleanup(func() {
		_ = ic.Close()
	})

	// Start the relayer on only the path between chainA<>chainB so that the initial transfer succeeds
	err = r.StartRelayer(ctx, eRep, pathAB)
	require.NoError(t, err)

	t.Cleanup(
		func() {
			err := r.StopRelayer(ctx, eRep)
			if err != nil {
				t.Logf("an error occured while stopping the relayer: %s", err)
			}
		},
	)

	// Fund user accounts with initial balances and get the transfer channel information between each set of chains
	initBal := math.NewInt(10_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), initBal, chainA, chainB, chainC)

	abChan, err := ibc.GetTransferChannel(ctx, r, eRep, chainIdA, chainIdB)
	require.NoError(t, err)

	baChan := abChan.Counterparty

	cbChan, err := ibc.GetTransferChannel(ctx, r, eRep, chainIdC, chainIdB)
	require.NoError(t, err)

	bcChan := cbChan.Counterparty

	userA, userB, userC := users[0], users[1], users[2]

	// Compose the prefixed denoms and ibc denom for asserting balances
	firstHopDenom := transfertypes.GetPrefixedDenom(baChan.PortID, baChan.ChannelID, chainA.Config().Denom)
	secondHopDenom := transfertypes.GetPrefixedDenom(cbChan.PortID, cbChan.ChannelID, firstHopDenom)

	firstHopDenomTrace := transfertypes.ParseDenomTrace(firstHopDenom)
	secondHopDenomTrace := transfertypes.ParseDenomTrace(secondHopDenom)

	firstHopIBCDenom := firstHopDenomTrace.IBCDenom()
	secondHopIBCDenom := secondHopDenomTrace.IBCDenom()

	firstHopEscrowAccount := transfertypes.GetEscrowAddress(abChan.PortID, abChan.ChannelID).String()
	secondHopEscrowAccount := transfertypes.GetEscrowAddress(bcChan.PortID, bcChan.ChannelID).String()

	zeroBal := math.ZeroInt()
	transferAmount := math.NewInt(100_000)

	// Attempt to send packet from Chain A->Chain B->Chain C->Chain D
	transfer := ibc.WalletAmount{
		Address: userB.FormattedAddress(),
		Denom:   chainA.Config().Denom,
		Amount:  transferAmount,
	}

	retries := uint8(10)

	firstHopMetadata := &PacketMetadata{
		Forward: &ForwardMetadata{
			Receiver: userC.FormattedAddress(),
			Channel:  bcChan.ChannelID,
			Port:     bcChan.PortID,
			Retries:  &retries,
			Timeout:  time.Second * 1, // Set low timeout for forward from chainB<>chainC
		},
	}

	memo, err := json.Marshal(firstHopMetadata)
	require.NoError(t, err)

	opts := ibc.TransferOptions{
		Memo: string(memo),
	}

	chainBHeight, err := chainB.Height(ctx)
	require.NoError(t, err)

	transferTx, err := chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName(), transfer, opts)
	require.NoError(t, err)

	// Poll for MsgRecvPacket on chainB
	_, err = cosmos.PollForMessage[*chantypes.MsgRecvPacket](ctx, chainB, cosmos.DefaultEncoding().InterfaceRegistry, chainBHeight, chainBHeight+20, nil)
	require.NoError(t, err)

	// Stop the relayer and wait for the timeout to happen on chainC
	err = r.StopRelayer(ctx, eRep)
	require.NoError(t, err)

	time.Sleep(time.Second * 11)

	// Restart the relayer
	err = r.StartRelayer(ctx, eRep, pathAB, pathBC)
	require.NoError(t, err)

	chainAHeight, err := chainA.Height(ctx)
	require.NoError(t, err)

	chainBHeight, err = chainB.Height(ctx)
	require.NoError(t, err)

	// Poll for the MsgTimeout on chainB and the MsgAck on chainA
	_, err = cosmos.PollForMessage[*chantypes.MsgTimeout](ctx, chainB, chainB.Config().EncodingConfig.InterfaceRegistry, chainBHeight, chainBHeight+20, nil)
	require.NoError(t, err)

	_, err = testutil.PollForAck(ctx, chainA, chainAHeight, chainAHeight+50, transferTx.Packet)
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 1, chainA)
	require.NoError(t, err)

	// Assert balances to ensure that the funds are still on the original sending chain
	chainABalance, err := chainA.GetBalance(ctx, userA.FormattedAddress(), chainA.Config().Denom)
	require.NoError(t, err)

	chainBBalance, err := chainB.GetBalance(ctx, userB.FormattedAddress(), firstHopIBCDenom)
	require.NoError(t, err)

	chainCBalance, err := chainC.GetBalance(ctx, userC.FormattedAddress(), secondHopIBCDenom)
	require.NoError(t, err)

	require.True(t, chainABalance.Equal(initBal))
	require.True(t, chainBBalance.Equal(zeroBal))
	require.True(t, chainCBalance.Equal(zeroBal))

	firstHopEscrowBalance, err := chainA.GetBalance(ctx, firstHopEscrowAccount, chainA.Config().Denom)
	require.NoError(t, err)

	secondHopEscrowBalance, err := chainB.GetBalance(ctx, secondHopEscrowAccount, firstHopIBCDenom)
	require.NoError(t, err)

	require.True(t, firstHopEscrowBalance.Equal(zeroBal))
	require.True(t, secondHopEscrowBalance.Equal(zeroBal))

	// Wait for blocks
	err = testutil.WaitForBlocks(ctx, 10, chainA, chainB, chainC)
	require.NoError(t, err)

	// loop through chain a -> chain d
	cosmosChains := []*cosmos.CosmosChain{chainA, chainB, chainC}
	for i := 0; i < len(cosmosChains); i++ {
		chain := cosmosChains[i]
		chain.StopAllNodes(ctx)

		// get exported packet forward middleware state
		stdOut, _, err := chain.GetNode().ExecBin(ctx, "export", "--modules-to-export=packetfowardmiddleware")
		require.NoError(t, err)

		chain.StartAllNodes(ctx)

		// validate that there are no in-flight packets
		var result PFMExport
		err = json.Unmarshal(stdOut, &result)
		require.NoError(t, err)
		require.Len(t, result.AppState.PacketForwardMiddleware.InFlightPackets, 0)
	}
}
