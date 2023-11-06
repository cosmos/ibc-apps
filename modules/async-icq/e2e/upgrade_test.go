package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	cosmosproto "github.com/cosmos/gogoproto/proto"
	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v7/types"
	"github.com/docker/docker/client"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/require"
)

const (
	chainName   = "simapp"
	upgradeName = "v2" // x/params migration

	haltHeightDelta    = uint64(9) // will propose upgrade this many blocks in the future
	blocksAfterUpgrade = uint64(7)

	VotingPeriod     = "15s"
	MaxDepositPeriod = "10s"
)

var (
	// baseChain is the current version of the chain that will be upgraded from
	// docker image load -i ../prev_builds/icq-host_7_0_0.tar
	baseChain = ibc.DockerImage{
		Repository: "icq-host",
		Version:    "v7.0.0",
		UidGid:     "1025:1025",
	}

	// make local-image
	upgradeTo = ibc.DockerImage{
		Repository: "icq-host",
		Version:    "local",
	}
)

func TestICQUpgrade(t *testing.T) {
	CosmosChainUpgradeTest(t, chainName, upgradeTo.Repository, upgradeTo.Version, upgradeName)
}

func CosmosChainUpgradeTest(t *testing.T, chainName, upgradeRepo, upgradeDockerTag, upgradeName string) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	t.Parallel()

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
		{
			Key:   "app_state.interchainquery.params.host_enabled",
			Value: false,
		},
		{
			Key:   "app_state.interchainquery.params.allow_queries",
			Value: []string{"/cosmos.bank.v1beta1.Query/AllBalances"},
		},
	}

	// Upgrade default to use the base chain image
	cfg := DefaultConfig
	cfg.ModifyGenesis = cosmos.ModifyGenesis(previousVersionGenesis)
	cfg.Images = []ibc.DockerImage{baseChain}

	numVals, numNodes := 2, 0
	chains := interchaintest.CreateChainWithConfig(t, numVals, numNodes, chainName, upgradeDockerTag, cfg)
	chain := chains[0].(*cosmos.CosmosChain)

	ctx, ic, client, _ := interchaintest.BuildInitialChain(t, chains, false)
	t.Cleanup(func() {
		ic.Close()
	})

	const userFunds = int64(10_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, chain)
	chainUser := users[0]

	// upgrade
	height, err := chain.Height(ctx)
	require.NoError(t, err, "error fetching height before submit upgrade proposal")

	haltHeight := height + haltHeightDelta
	proposalID := SubmitUpgradeProposal(t, ctx, chain, chainUser, upgradeName, haltHeight)

	ValidatorVoting(t, ctx, chain, proposalID, height, haltHeight)
	UpgradeNodes(t, ctx, chain, client, haltHeight, upgradeRepo, upgradeDockerTag)

	// Validate the ICQ subspace -> keeper migration was successful.
	cmd := []string{
		chain.Config().Bin, "q", "interchainquery", "params", "--output=json", "--node", chain.GetRPCAddress(),
	}
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	fmt.Println("stdout", string(stdout))
	require.NoError(t, err, "error fetching icq params")

	var params icqtypes.Params
	err = json.Unmarshal(stdout, &params)
	require.NoError(t, err, "error unmarshalling icq params")

	t.Logf("params: %+v", params)
	require.Equal(t, false, params.HostEnabled, "HostEnabled not equal to expected value")
	require.Equal(t, []string{"/cosmos.bank.v1beta1.Query/AllBalances"}, params.AllowQueries, "AllowQueries not equal to expected value")

}

func SubmitUpgradeProposal(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, upgradeName string, haltHeight uint64) string {
	upgradeMsg := []cosmosproto.Message{
		&upgradetypes.MsgSoftwareUpgrade{
			// Gov Module account
			Authority: "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
			Plan: upgradetypes.Plan{
				Name:   upgradeName,
				Height: int64(haltHeight),
			},
		},
	}

	proposal, err := chain.BuildProposal(upgradeMsg, "Chain Upgrade "+upgradeName, "Summary desc", "ipfs://CID", fmt.Sprintf(`500000000%s`, chain.Config().Denom))
	require.NoError(t, err, "error building proposal")

	txProp, err := chain.SubmitProposal(ctx, user.KeyName(), proposal)
	require.NoError(t, err, "error submitting proposal")

	t.Log("txProp", txProp)
	return txProp.ProposalID
}

func UpgradeNodes(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, client *client.Client, haltHeight uint64, upgradeRepo, upgradeBranchVersion string) {
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

func ValidatorVoting(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, proposalID string, height uint64, haltHeight uint64) {
	err := chain.VoteOnProposalAllValidators(ctx, proposalID, cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	_, err = cosmos.PollForProposalStatus(ctx, chain, height, height+haltHeightDelta, proposalID, cosmos.ProposalStatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

	timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*45)
	defer timeoutCtxCancel()

	height, err = chain.Height(ctx)
	require.NoError(t, err, "error fetching height before upgrade")

	// this should timeout due to chain halt at upgrade height.
	_ = testutil.WaitForBlocks(timeoutCtx, int(haltHeight-height), chain)

	height, err = chain.Height(ctx)
	require.NoError(t, err, "error fetching height after chain should have halted")

	// make sure that chain is halted
	require.Equal(t, haltHeight, height, "height is not equal to halt height")
}
