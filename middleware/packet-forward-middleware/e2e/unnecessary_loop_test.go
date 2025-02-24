package e2e

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"cosmossdk.io/math"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/relayer"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TestUnnecessaryLoop tests that a packet that is sent from ChainA -> ChainB -> ChainA
// with failure on second hop will properly refund and have proper escrow balances.
func TestUnnecessaryLoop(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	var (
		ctx                = context.Background()
		client, network    = interchaintest.DockerSetup(t)
		rep                = testreporter.NewNopReporter()
		eRep               = rep.RelayerExecReporter(t)
		chainIdA, chainIdB = "chain-a", "chain-b"
	)

	vals := 1
	fullNodes := 0

	baseCfg := DefaultConfig

	baseCfg.ChainID = chainIdA
	configA := baseCfg

	baseCfg.ChainID = chainIdB
	configB := baseCfg

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{Name: "pfm", ChainConfig: configA, NumFullNodes: &fullNodes, NumValidators: &vals},
		{Name: "pfm", ChainConfig: configB, NumFullNodes: &fullNodes, NumValidators: &vals},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	chainA, chainB := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)

	r := interchaintest.NewBuiltinRelayerFactory(
		ibc.Hermes,
		zaptest.NewLogger(t),
		relayer.DockerImage(&DefaultRelayer),
	).Build(t, client, network)

	const pathAB = "ab"

	ic := interchaintest.NewInterchain().
		AddChain(chainA).
		AddChain(chainB).
		AddRelayer(r, "relayer").
		AddLink(interchaintest.InterchainLink{
			Chain1:  chainA,
			Chain2:  chainB,
			Relayer: r,
			Path:    pathAB,
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
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), initBal, chainA, chainB)

	abChan, err := ibc.GetTransferChannel(ctx, r, eRep, chainIdA, chainIdB)
	require.NoError(t, err)

	baChan := abChan.Counterparty

	userA, userB := users[0], users[1]

	// Compose the prefixed denoms and ibc denom for asserting balances
	firstHopDenom := transfertypes.GetPrefixedDenom(baChan.PortID, baChan.ChannelID, chainA.Config().Denom)
	firstHopDenomTrace := transfertypes.ParseDenomTrace(firstHopDenom)
	firstHopIBCDenom := firstHopDenomTrace.IBCDenom()
	firstHopEscrowAccount := transfertypes.GetEscrowAddress(abChan.PortID, abChan.ChannelID).String()

	zeroBal := math.ZeroInt()
	transferAmount := math.NewInt(100_000)

	// Attempt to send packet from Chain A->Chain B->Chain C->Chain D
	transfer := ibc.WalletAmount{
		Address: userB.FormattedAddress(),
		Denom:   chainA.Config().Denom,
		Amount:  transferAmount,
	}

	retries := uint8(0)

	firstHopMetadata := &PacketMetadata{
		Forward: &ForwardMetadata{
			Receiver: "malformed address",
			Channel:  baChan.ChannelID,
			Port:     baChan.PortID,
			Retries:  &retries,
			Timeout:  time.Second * 100,
		},
	}

	memo, err := json.Marshal(firstHopMetadata)
	require.NoError(t, err)

	opts := ibc.TransferOptions{
		Memo: string(memo),
	}

	transferTx, err := chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName(), transfer, opts)
	require.NoError(t, err)

	chainAHeight, err := chainA.Height(ctx)
	require.NoError(t, err)

	_, err = testutil.PollForAck(ctx, chainA, chainAHeight, chainAHeight+30, transferTx.Packet)
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 1, chainA)
	require.NoError(t, err)

	// Assert balances to ensure that the funds are still on the original sending chain
	chainABalance, err := chainA.GetBalance(ctx, userA.FormattedAddress(), chainA.Config().Denom)
	require.NoError(t, err)

	chainBBalance, err := chainB.GetBalance(ctx, userB.FormattedAddress(), firstHopIBCDenom)
	require.NoError(t, err)

	require.True(t, chainABalance.Equal(initBal))
	require.True(t, chainBBalance.Equal(zeroBal))

	firstHopEscrowBalance, err := chainA.GetBalance(ctx, firstHopEscrowAccount, chainA.Config().Denom)
	require.NoError(t, err)
	require.True(t, firstHopEscrowBalance.Equal(zeroBal))
}
