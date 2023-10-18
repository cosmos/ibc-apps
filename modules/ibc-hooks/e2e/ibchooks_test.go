package e2e

import (
	"fmt"
	"strings"
	"testing"

	"cosmossdk.io/math"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/require"
)

// TestIBCHooks ensures the ibc-hooks middleware from osmosis works as expected.
func TestIBCHooks(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	// Create chain factory with osmosis and osmosis2
	numVals := 1
	numFullNodes := 0

	genesisWalletAmount := int64(10_000_000)

	cfg := ibc.ChainConfig{
		Name:    "osmosis",
		Type:    "cosmos",
		ChainID: "simapp-1",
		Bin:     "simd",
		Images: []ibc.DockerImage{
			{
				Repository: "ibchooks",
				Version:    "local",
				UidGid:     "1025:1025",
			},
		},
		Bech32Prefix:   "cosmos",
		Denom:          "uosmo",
		CoinType:       "118",
		GasPrices:      "0uosmo",
		GasAdjustment:  1.5,
		TrustingPeriod: "330h",
		EncodingConfig: WasmEncodingConfig(),
	}

	cosmos.SetSDKConfig(cfg.Bech32Prefix)

	cfg2 := cfg.Clone()
	cfg2.Name = "counterparty"
	cfg2.ChainID = "counterparty-2"

	chains := interchaintest.CreateChainsWithChainSpecs(t, []*interchaintest.ChainSpec{
		{
			Name:          "osmosis",
			ChainName:     "osmosis",
			ChainConfig:   cfg,
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
		{
			Name:          "counterparty",
			ChainName:     "counterparty",
			ChainConfig:   cfg2,
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
	})

	osmosis, osmosis2 := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)

	const path = "ibc-path"
	enableBlockDB := false
	skipPathCreations := false
	ctx, _, r, _, eRep, _, _ := interchaintest.BuildInitialChainWithRelayer(
		t,
		chains,
		enableBlockDB,
		ibc.CosmosRly,
		[]string{"--processor", "events", "--block-history", "100"},
		[]interchaintest.InterchainLink{
			{
				Chain1: osmosis,
				Chain2: osmosis2,
				Path:   path,
			},
		},
		skipPathCreations,
	)

	if err := r.StartRelayer(ctx, eRep, path); err != nil {
		t.Fatal(err)
	}

	// Create some user accounts on both chains
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), genesisWalletAmount, osmosis, osmosis2)

	// Get our Bech32 encoded user addresses
	osmosisUser, osmosis2User := users[0], users[1]

	osmosisUserAddr := osmosisUser.FormattedAddress()
	// osmosis2UserAddr := osmosis2User.FormattedAddress()

	channel, err := ibc.GetTransferChannel(ctx, r, eRep, osmosis.Config().ChainID, osmosis2.Config().ChainID)
	require.NoError(t, err)

	_, contractAddr := SetupContract(t, ctx, osmosis2, osmosis2User.KeyName(), "contracts/ibchooks_counter.wasm", `{"count":0}`)

	// do an ibc transfer through the memo to the other chain.
	transfer := ibc.WalletAmount{
		Address: contractAddr,
		Denom:   osmosis.Config().Denom,
		Amount:  math.NewInt(1),
	}

	memo := ibc.TransferOptions{
		Memo: fmt.Sprintf(`{"wasm":{"contract":"%s","msg":%s}}`, contractAddr, `{"increment":{}}`),
	}

	// Initial transfer. Account is created by the wasm execute is not so we must do this twice to properly set up
	transferTx, err := osmosis.SendIBCTransfer(ctx, channel.ChannelID, osmosisUser.KeyName(), transfer, memo)
	require.NoError(t, err)

	osmosisHeight, err := osmosis.Height(ctx)
	require.NoError(t, err)

	_, err = testutil.PollForAck(ctx, osmosis, osmosisHeight-5, osmosisHeight+25, transferTx.Packet)
	require.NoError(t, err)

	// Second time, this will make the counter == 1 since the account is now created.
	transferTx, err = osmosis.SendIBCTransfer(ctx, channel.ChannelID, osmosisUser.KeyName(), transfer, memo)
	require.NoError(t, err)
	osmosisHeight, err = osmosis.Height(ctx)
	require.NoError(t, err)

	_, err = testutil.PollForAck(ctx, osmosis, osmosisHeight-5, osmosisHeight+25, transferTx.Packet)
	require.NoError(t, err)

	// Get the address on the other chain's side
	addr := GetIBCHooksUserAddress(t, ctx, osmosis, channel.ChannelID, osmosisUserAddr)
	require.NotEmpty(t, addr)

	fmt.Println(transferTx)
	fmt.Println("Waiting for blocks...", osmosis.GetNode().Chain.GetHostRPCAddress(), osmosis2.GetNode().Chain.GetHostRPCAddress())
	fmt.Println("userAddr", osmosisUserAddr, "uAddr", addr)
	fmt.Println("contractAddr", contractAddr)
	// simd query ibchooks wasm-sender channel-0 cosmos1d6689gwhh52ld4aysh2qw7ms7j6umw6gpynnv2 --node=http://127.0.0.1:40027
	// simd q wasm contract-state smart cosmos14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s4hmalr '{"get_total_funds":{"addr":"cosmos12ufqv9hkh07kznkzal4l0mg86zjzduk0ddxkx68fkkszz9djglnsa0a2vc"}}' --node=http://127.0.0.1:41655
	// TODO: testutil.WaitForBlocks(ctx, 50000, osmosis)

	// Get funds on the receiving chain
	funds := GetIBCHookTotalFunds(t, ctx, osmosis2, contractAddr, addr)
	require.Equal(t, int(1), len(funds.Data.TotalFunds))

	var ibcDenom string
	for _, coin := range funds.Data.TotalFunds {
		if strings.HasPrefix(coin.Denom, "ibc/") {
			ibcDenom = coin.Denom
			break
		}
	}
	require.NotEmpty(t, ibcDenom)

	// ensure the count also increased to 1 as expected.
	count := GetIBCHookCount(t, ctx, osmosis2, contractAddr, addr)
	require.Equal(t, int64(1), count.Data.Count)
}
