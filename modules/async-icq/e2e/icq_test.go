package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/icza/dyno"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/relayer"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TestInterchainQueries spins up a controller and host chain, using a demo controller implementation,
// and asserts that a bank query can successfully be executed on the host chain and the results can be
// retrieved on the controller chain.
// See: https://github.com/strangelove-ventures/interchain-query-demo
func TestInterchainQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	t.Parallel()

	client, network := interchaintest.DockerSetup(t)

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	ctx := context.Background()

	numVals := 1
	numNodes := 0

	dockerImage := ibc.DockerImage{
		Repository: "ghcr.io/strangelove-ventures/heighliner/icqd",
		Version:    "latest",
		UidGid:     "1025:1025",
	}

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			ChainName:     "controller",
			NumValidators: &numVals,
			NumFullNodes:  &numNodes,
			ChainConfig: ibc.ChainConfig{
				Type:           "cosmos",
				Name:           "controller",
				ChainID:        "controller",
				Images:         []ibc.DockerImage{dockerImage},
				Bin:            "icq",
				Bech32Prefix:   "cosmos",
				Denom:          "atom",
				GasPrices:      "0.00atom",
				TrustingPeriod: "300h",
				GasAdjustment:  1.1,
			}},
		{
			ChainName:     "host",
			NumValidators: &numVals,
			NumFullNodes:  &numNodes,
			ChainConfig: ibc.ChainConfig{
				Type:           "cosmos",
				Name:           "host",
				ChainID:        "host",
				Images:         []ibc.DockerImage{dockerImage},
				Bin:            "icq",
				Bech32Prefix:   "cosmos",
				Denom:          "atom",
				GasPrices:      "0.00atom",
				TrustingPeriod: "300h",
				GasAdjustment:  1.1,
				ModifyGenesis:  modifyGenesisAllowICQQueries([]string{"/cosmos.bank.v1beta1.Query/AllBalances"}), // Add the whitelisted queries to the host chain
			}},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	controllerChain, hostChain := chains[0], chains[1]

	r := interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(t),
		relayer.StartupFlags("-b", "100"),
	).Build(t, client, network)

	const pathName = "host-controller"
	const relayerName = "relayer"

	ic := interchaintest.NewInterchain().
		AddChain(controllerChain).
		AddChain(hostChain).
		AddRelayer(r, relayerName).
		AddLink(interchaintest.InterchainLink{
			Chain1:  controllerChain,
			Chain2:  hostChain,
			Relayer: r,
			Path:    pathName,
			CreateChannelOpts: ibc.CreateChannelOptions{
				SourcePortName: "interquery",
				DestPortName:   "icqhost",
				Order:          ibc.Unordered,
				Version:        "icq-1",
			},
		})

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:  t.Name(),
		Client:    client,
		NetworkID: network,

		SkipPathCreation: false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	// Fund user accounts, so we can query balances and make assertions.
	const userFunds = int64(10_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, controllerChain, hostChain)
	controllerUser := users[0]
	hostUser := users[1]

	// Wait a few blocks for user accounts to be created on chain.
	err = testutil.WaitForBlocks(ctx, 5, controllerChain, hostChain)
	require.NoError(t, err)

	// Query for the recently created channel-id.
	channels, err := r.GetChannels(ctx, eRep, controllerChain.Config().ChainID)
	require.NoError(t, err)

	// Start the relayer.
	err = r.StartRelayer(ctx, eRep, pathName)
	require.NoError(t, err)

	t.Cleanup(
		func() {
			err := r.StopRelayer(ctx, eRep)
			if err != nil {
				t.Logf("an error occured while stopping the relayer: %s", err)
			}
		},
	)

	// Wait a few blocks for the relayer to start.
	err = testutil.WaitForBlocks(ctx, 5, controllerChain, hostChain)
	require.NoError(t, err)

	// Query for the balances of an account on the counterparty chain using interchain queries.
	chanID := channels[0].Counterparty.ChannelID
	require.NotEmpty(t, chanID)

	controllerAddr := controllerUser.(*cosmos.CosmosWallet).FormattedAddress()
	require.NotEmpty(t, controllerAddr)

	hostAddr := hostUser.(*cosmos.CosmosWallet).FormattedAddress()
	require.NotEmpty(t, hostAddr)

	cmd := []string{"icq", "tx", "interquery", "send-query-all-balances", chanID, hostAddr,
		"--node", controllerChain.GetRPCAddress(),
		"--home", controllerChain.HomeDir(),
		"--chain-id", controllerChain.Config().ChainID,
		"--from", controllerAddr,
		"--keyring-dir", controllerChain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}
	_, _, err = controllerChain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	// Wait a few blocks for query to be sent to counterparty.
	err = testutil.WaitForBlocks(ctx, 5, controllerChain)
	require.NoError(t, err)

	// Check the results from the interchain query above.
	cmd = []string{"icq", "query", "interquery", "query-state", "1",
		"--node", controllerChain.GetRPCAddress(),
		"--home", controllerChain.HomeDir(),
		"--chain-id", controllerChain.Config().ChainID,
		"--output", "json",
	}
	stdout, _, err := controllerChain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	results := &icqResults{}
	err = json.Unmarshal(stdout, results)
	require.NoError(t, err)
	require.NotEmpty(t, results.Request)
	require.NotEmpty(t, results.Response)
}

type icqResults struct {
	Request struct {
		Type       string `json:"@type"`
		Address    string `json:"address"`
		Pagination struct {
			Key        interface{} `json:"key"`
			Offset     string      `json:"offset"`
			Limit      string      `json:"limit"`
			CountTotal bool        `json:"count_total"`
			Reverse    bool        `json:"reverse"`
		} `json:"pagination"`
	} `json:"request"`
	Response struct {
		Type     string `json:"@type"`
		Balances []struct {
			Amount string `json:"amount"`
			Denom  string `json:"denom"`
		} `json:"balances"`
		Pagination struct {
			NextKey interface{} `json:"next_key"`
			Total   string      `json:"total"`
		} `json:"pagination"`
	} `json:"response"`
}

func modifyGenesisAllowICQQueries(allowQueries []string) func(ibc.ChainConfig, []byte) ([]byte, error) {
	return func(chainConfig ibc.ChainConfig, genbz []byte) ([]byte, error) {
		g := make(map[string]interface{})
		if err := json.Unmarshal(genbz, &g); err != nil {
			return nil, fmt.Errorf("failed to unmarshal genesis file: %w", err)
		}
		if err := dyno.Set(g, allowQueries, "app_state", "interchainquery", "params", "allow_queries"); err != nil {
			return nil, fmt.Errorf("failed to set allowed interchain queries in genesis json: %w", err)
		}
		out, err := json.Marshal(g)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal genesis bytes to json: %w", err)
		}
		return out, nil
	}
}
