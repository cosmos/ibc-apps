package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	"github.com/docker/docker/client"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/relayer"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	chainName   = "simapp"
	upgradeName = "v3" // escrow state and balance re-sync upgrade

	haltHeightDelta    = int64(20) // will propose upgrade this many blocks in the future
	blocksAfterUpgrade = int64(7)

	VotingPeriod     = "15s"
	MaxDepositPeriod = "10s"
)

var (
	// baseChain is the current version of the chain that will be upgraded from
	// docker image load -i ../prev_builds/pfm_8_1_0.tar
	baseChain = ibc.DockerImage{
		Repository: "pfm",
		Version:    "v8.1.0",
		UIDGID:     "1025:1025",
	}

	// make local-image
	upgradeTo = ibc.DockerImage{
		Repository: "pfm",
		Version:    "local",
	}
)

func TestPFMUpgrade(t *testing.T) {
	CosmosChainUpgradeTest(t, chainName, upgradeTo.Repository, upgradeTo.Version, upgradeName)
}

func CosmosChainUpgradeTest(t *testing.T, chainName, upgradeRepo, upgradeDockerTag, upgradeName string) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	previousVersionGenesis := []cosmos.GenesisKV{
		{
			Key:   "app_state.gov.params.voting_period",
			Value: VotingPeriod,
		},
		{
			Key:   "app_state.gov.params.max_deposit_period",
			Value: MaxDepositPeriod,
		},
		{
			Key:   "app_state.gov.params.min_deposit.0.denom",
			Value: Denom,
		},
	}

	var (
		ctx                                    = context.Background()
		client, network                        = interchaintest.DockerSetup(t)
		rep                                    = testreporter.NewNopReporter()
		eRep                                   = rep.RelayerExecReporter(t)
		chainIdA, chainIdB, chainIdC, chainIdD = "chain-1", "chain-2", "chain-3", "chain-4"
		waitBlocks                             = 3
	)

	cfgA := DefaultConfig
	cfgA.ChainID = chainIdA

	// ChainB is the chain that will be upgraded
	cfgB := DefaultConfig
	cfgB.ModifyGenesis = cosmos.ModifyGenesis(previousVersionGenesis)
	cfgB.Images = []ibc.DockerImage{baseChain}
	cfgB.ChainID = chainIdB

	cfgC := DefaultConfig
	cfgC.ChainID = chainIdC

	cfgD := DefaultConfig
	cfgD.ChainID = chainIdD

	numValsPrimary, numNodes := 2, 0
	numVals := 1

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{Name: "pfm", ChainConfig: cfgA, NumFullNodes: &numNodes, NumValidators: &numVals},
		{Name: "pfm", ChainConfig: cfgB, NumFullNodes: &numNodes, NumValidators: &numValsPrimary},
		{Name: "pfm", ChainConfig: cfgC, NumFullNodes: &numNodes, NumValidators: &numVals},
		{Name: "pfm", ChainConfig: cfgD, NumFullNodes: &numNodes, NumValidators: &numVals},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	chainA, chainB, chainC, chainD := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain), chains[2].(*cosmos.CosmosChain), chains[3].(*cosmos.CosmosChain)

	r := interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(t),
		relayer.DockerImage(&DefaultRelayer),
	).Build(t, client, network)

	const pathAB = "ab"
	const pathBC = "bc"
	const pathCD = "cd"

	ic := interchaintest.NewInterchain().
		AddChain(chainA).
		AddChain(chainB).
		AddChain(chainC).
		AddChain(chainD).
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
		}).
		AddLink(interchaintest.InterchainLink{
			Chain1:  chainC,
			Chain2:  chainD,
			Relayer: r,
			Path:    pathCD,
		})

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:          t.Name(),
		Client:            client,
		NetworkID:         network,
		BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),

		SkipPathCreation: false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	initBal := math.NewInt(10_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), initBal, chainA, chainB, chainC, chainD)

	// -------------------------------------------------------------------------
	// IBC setup

	abChan, err := ibc.GetTransferChannel(ctx, r, eRep, chainIdA, chainIdB)
	require.NoError(t, err)

	baChan := abChan.Counterparty

	cbChan, err := ibc.GetTransferChannel(ctx, r, eRep, chainIdC, chainIdB)
	require.NoError(t, err)

	bcChan := cbChan.Counterparty

	dcChan, err := ibc.GetTransferChannel(ctx, r, eRep, chainIdD, chainIdC)
	require.NoError(t, err)

	cdChan := dcChan.Counterparty

	// Start the relayer on both paths
	err = r.StartRelayer(ctx, eRep, pathAB, pathBC, pathCD)
	require.NoError(t, err)

	t.Cleanup(
		func() {
			err := r.StopRelayer(ctx, eRep)
			if err != nil {
				t.Logf("an error occurred while stopping the relayer: %s", err)
			}
		},
	)

	// Get original account balances
	userA, userB, userC, userD := users[0], users[1], users[2], users[3]

	// Compose the prefixed denoms and ibc denom for asserting balances
	// firstHopDenom := transfertypes.GetPrefixedDenom(baChan.PortID, baChan.ChannelID, chainA.Config().Denom)
	// secondHopDenom := transfertypes.GetPrefixedDenom(cbChan.PortID, cbChan.ChannelID, firstHopDenom)
	// thirdHopDenom := transfertypes.GetPrefixedDenom(dcChan.PortID, dcChan.ChannelID, secondHopDenom)

	// firstHopDenomTrace := transfertypes.ParseDenomTrace(firstHopDenom)
	// secondHopDenomTrace := transfertypes.ParseDenomTrace(secondHopDenom)
	// thirdHopDenomTrace := transfertypes.ParseDenomTrace(thirdHopDenom)

	// firstHopIBCDenom := firstHopDenomTrace.IBCDenom()
	// secondHopIBCDenom := secondHopDenomTrace.IBCDenom()
	// thirdHopIBCDenom := thirdHopDenomTrace.IBCDenom()

	// firstHopEscrowAccount := sdk.MustBech32ifyAddressBytes(chainA.Config().Bech32Prefix, transfertypes.GetEscrowAddress(abChan.PortID, abChan.ChannelID))
	// secondHopEscrowAccount := sdk.MustBech32ifyAddressBytes(chainB.Config().Bech32Prefix, transfertypes.GetEscrowAddress(bcChan.PortID, bcChan.ChannelID))
	// thirdHopEscrowAccount := sdk.MustBech32ifyAddressBytes(chainC.Config().Bech32Prefix, transfertypes.GetEscrowAddress(cdChan.PortID, abChan.ChannelID))

	zeroBal := math.ZeroInt()
	transferAmount := math.NewInt(100_000)

	// -------------------------------------------------------------------------
	// Same as ./packet_forward_test.go "multi-hop through native chain ack error refund"
	// to reproduce the invalid state

	// send normal IBC transfer from B->A to get funds in IBC denom, then do multihop A->B(native)->C->D
	// this lets us test the burn from escrow account on chain C and the escrow to escrow transfer on chain B.

	// Compose the prefixed denoms and ibc denom for asserting balances
	baDenom := transfertypes.GetPrefixedDenom(abChan.PortID, abChan.ChannelID, chainB.Config().Denom)
	bcDenom := transfertypes.GetPrefixedDenom(cbChan.PortID, cbChan.ChannelID, chainB.Config().Denom)
	cdDenom := transfertypes.GetPrefixedDenom(dcChan.PortID, dcChan.ChannelID, bcDenom)

	baDenomTrace := transfertypes.ParseDenomTrace(baDenom)
	bcDenomTrace := transfertypes.ParseDenomTrace(bcDenom)
	cdDenomTrace := transfertypes.ParseDenomTrace(cdDenom)

	baIBCDenom := baDenomTrace.IBCDenom()
	bcIBCDenom := bcDenomTrace.IBCDenom()
	cdIBCDenom := cdDenomTrace.IBCDenom()

	transfer := ibc.WalletAmount{
		Address: userA.FormattedAddress(),
		Denom:   chainB.Config().Denom,
		Amount:  transferAmount,
	}

	chainBHeight, err := chainB.Height(ctx)
	require.NoError(t, err)

	transferTx, err := chainB.SendIBCTransfer(ctx, baChan.ChannelID, userB.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err)
	_, err = testutil.PollForAck(ctx, chainB, chainBHeight, chainBHeight+10, transferTx.Packet)
	require.NoError(t, err)
	err = testutil.WaitForBlocks(ctx, waitBlocks, chainB)
	require.NoError(t, err)

	// assert balance for user controlled wallet
	chainABalance, err := chainA.GetBalance(ctx, userA.FormattedAddress(), baIBCDenom)
	require.NoError(t, err)

	baEscrowBalance, err := chainB.GetBalance(ctx, transfertypes.GetEscrowAddress(baChan.PortID, baChan.ChannelID).String(), chainB.Config().Denom)
	require.NoError(t, err)

	require.True(t, chainABalance.Equal(transferAmount))
	require.True(t, baEscrowBalance.Equal(transferAmount))

	// Send a malformed packet with invalid receiver address from Chain A->Chain B->Chain C->Chain D
	// This should succeed in the first hop and second hop, then fail to make the third hop.
	// Funds should be refunded to Chain B and then to Chain A via acknowledgements with errors.
	transfer = ibc.WalletAmount{
		Address: userB.FormattedAddress(),
		Denom:   baIBCDenom,
		Amount:  transferAmount,
	}

	secondHopMetadata := &PacketMetadata{
		Forward: &ForwardMetadata{
			Receiver: "xyz1t8eh66t2w5k67kwurmn5gqhtq6d2ja0vp7jmmq", // malformed receiver address on chain D
			Channel:  cdChan.ChannelID,
			Port:     cdChan.PortID,
		},
	}

	nextBz, err := json.Marshal(secondHopMetadata)
	require.NoError(t, err)

	next := string(nextBz)

	firstHopMetadata := &PacketMetadata{
		Forward: &ForwardMetadata{
			Receiver: userC.FormattedAddress(),
			Channel:  bcChan.ChannelID,
			Port:     bcChan.PortID,
			Next:     &next,
		},
	}

	memo, err := json.Marshal(firstHopMetadata)
	require.NoError(t, err)

	chainAHeight, err := chainA.Height(ctx)
	require.NoError(t, err)

	transferTx, err = chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName(), transfer, ibc.TransferOptions{Memo: string(memo)})
	require.NoError(t, err)
	_, err = testutil.PollForAck(ctx, chainA, chainAHeight, chainAHeight+30, transferTx.Packet)
	require.NoError(t, err)
	err = testutil.WaitForBlocks(ctx, waitBlocks, chainA)
	require.NoError(t, err)

	// assert balances for user controlled wallets
	chainDBalance, err := chainD.GetBalance(ctx, userD.FormattedAddress(), cdIBCDenom)
	require.NoError(t, err)

	chainCBalance, err := chainC.GetBalance(ctx, userC.FormattedAddress(), bcIBCDenom)
	require.NoError(t, err)

	chainBBalance, err := chainB.GetBalance(ctx, userB.FormattedAddress(), chainB.Config().Denom)
	require.NoError(t, err)

	chainABalance, err = chainA.GetBalance(ctx, userA.FormattedAddress(), baIBCDenom)
	require.NoError(t, err)

	require.True(t, chainABalance.Equal(transferAmount))
	require.True(t, chainBBalance.Equal(initBal.Sub(transferAmount)))
	require.True(t, chainCBalance.Equal(zeroBal))
	require.True(t, chainDBalance.Equal(zeroBal))

	// assert balances for IBC escrow accounts
	bcEscrowBalance, err := chainB.GetBalance(ctx, transfertypes.GetEscrowAddress(bcChan.PortID, bcChan.ChannelID).String(), chainB.Config().Denom)
	require.NoError(t, err)

	baEscrowBalance, err = chainB.GetBalance(ctx, transfertypes.GetEscrowAddress(baChan.PortID, baChan.ChannelID).String(), chainB.Config().Denom)
	require.NoError(t, err)

	require.True(t, baEscrowBalance.Equal(transferAmount))
	require.True(t, bcEscrowBalance.Equal(zeroBal))

	conn, err := grpc.Dial(chainB.GetHostGRPCAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	queryClient := transfertypes.NewQueryClient(conn)

	req := &transfertypes.QueryTotalEscrowForDenomRequest{Denom: chainB.Config().Denom}
	res, err := queryClient.TotalEscrowForDenom(ctx, req)
	require.NoError(t, err)

	// assert the WRONG escrow state before the upgrade
	require.Falsef(t,
		baEscrowBalance.Equal(res.Amount.Amount),
		"before upgrade: expected B->A escrow amount %s to NOT equal reported B total escrow %s",
		baEscrowBalance.String(), res.Amount.Amount.String(),
	)

	// Assert baEscrowBalance was still decreased by the transfer
	require.True(t, baEscrowBalance.Equal(transferAmount))

	t.Logf("before: baEscrowBalance: %s", baEscrowBalance.String())
	t.Logf("before: total escrow for denom: %s", res.Amount.Amount.String())

	// -------------------------------------------------------------------------
	// ChainB upgrade, as it is the chain with the escrow state that will be corrected
	height, err := chainB.Height(ctx)
	require.NoError(t, err, "error fetching height before submit upgrade proposal")

	haltHeight := height + haltHeightDelta
	proposalID := SubmitUpgradeProposal(t, ctx, chainB, userB, upgradeName, haltHeight)

	ValidatorVoting(t, ctx, chainB, proposalID, height, haltHeight)
	UpgradeNodes(t, ctx, chainB, client, haltHeight, upgradeRepo, upgradeDockerTag)

	// Re-create client
	conn, err = grpc.Dial(chainB.GetHostGRPCAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	queryClient = transfertypes.NewQueryClient(conn)

	// Validate escrow state after migration matches escrow account balances
	reqAfter := &transfertypes.QueryTotalEscrowForDenomRequest{Denom: chainB.Config().Denom}
	totalEscrowAfter, err := queryClient.TotalEscrowForDenom(ctx, reqAfter)
	require.NoError(t, err)

	// assert the CORRECT escrow state after the upgrade
	require.Truef(
		t,
		baEscrowBalance.Equal(totalEscrowAfter.Amount.Amount),
		"after upgrade: expected B->A escrow amount %s to equal reported B total escrow %s in bank",
		baEscrowBalance.String(), totalEscrowAfter.Amount.Amount.String(),
	)

	t.Logf("after: baEscrowBalance: %s", baEscrowBalance.String())
	t.Logf("after: total escrow for denom: %s", totalEscrowAfter.Amount.Amount.String())
}

func SubmitUpgradeProposal(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, upgradeName string, haltHeight int64) string {
	upgradeMsg := []cosmos.ProtoMessage{
		&upgradetypes.MsgSoftwareUpgrade{
			// Gov Module account
			Authority: "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
			Plan: upgradetypes.Plan{
				Name:   upgradeName,
				Height: int64(haltHeight),
			},
		},
	}

	proposal, err := chain.BuildProposal(upgradeMsg, "Chain Upgrade "+upgradeName, "Summary desc", "ipfs://CID", fmt.Sprintf(`500000000%s`, chain.Config().Denom), user.KeyName(), false)
	require.NoError(t, err, "error building proposal")

	txProp, err := chain.SubmitProposal(ctx, user.KeyName(), proposal)
	require.NoError(t, err, "error submitting proposal")

	t.Log("txProp", txProp)
	return txProp.ProposalID
}

func UpgradeNodes(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, client *client.Client, haltHeight int64, upgradeRepo, upgradeBranchVersion string) {
	// bring down nodes to prepare for upgrade
	t.Log("stopping node(s)")
	err := chain.StopAllNodes(ctx)
	require.NoError(t, err, "error stopping node(s)")

	// upgrade version on all nodes
	t.Log("upgrading node(s)")
	chain.UpgradeVersion(ctx, client, upgradeRepo, upgradeBranchVersion)

	// start all nodes back up.
	// validators reach consensus on first block after upgrade height
	// and chain block production resumes.
	t.Log("starting node(s)")
	err = chain.StartAllNodes(ctx)
	require.NoError(t, err, "error starting upgraded node(s)")

	timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*60)
	defer timeoutCtxCancel()

	err = testutil.WaitForBlocks(timeoutCtx, int(blocksAfterUpgrade), chain)
	require.NoError(t, err, "chain did not produce blocks after upgrade")

	height, err := chain.Height(ctx)
	require.NoError(t, err, "error fetching height after upgrade")

	require.GreaterOrEqual(t, height, haltHeight+blocksAfterUpgrade, "height did not increment enough after upgrade")
}

func ValidatorVoting(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, proposalID string, height int64, haltHeight int64) {
	proposalInt, err := strconv.ParseUint(proposalID, 10, 64)
	require.NoError(t, err, "failed to parse proposalID")

	err = chain.VoteOnProposalAllValidators(ctx, proposalInt, cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	_, err = cosmos.PollForProposalStatusV1(ctx, chain, height, height+haltHeightDelta, proposalInt, govv1.ProposalStatus_PROPOSAL_STATUS_PASSED)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

	timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*45)
	defer timeoutCtxCancel()

	height, err = chain.Height(ctx)
	require.NoError(t, err, "error fetching height before upgrade")

	// this should timeout due to chain halt at upgrade height.
	_ = testutil.WaitForBlocks(timeoutCtx, int(haltHeight-height), chain)

	// Wait another 10 seconds after chain should have halted to ensure it is halted
	time.Sleep(time.Second * 10)

	height, err = chain.Height(ctx)
	require.NoError(t, err, "error fetching height after chain should have halted")

	// make sure that chain is halted
	require.Equal(t, haltHeight, height, "height is not equal to halt height, chain did not halt")
}
