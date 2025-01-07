package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v9/modules/apps/transfer/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/relayer"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"golang.org/x/sync/errgroup"
)

var chars = []byte("abcdefghijklmnopqrstuvwxyz")

// RandLowerCaseLetterString returns a lowercase letter string of given length
func RandLowerCaseLetterString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func BuildWallets(ctx context.Context, testName string, chains ...ibc.Chain) ([]ibc.Wallet, error) {
	users := make([]ibc.Wallet, len(chains))
	for i, chain := range chains {
		chainCfg := chain.Config()
		keyName := fmt.Sprintf("%s-%s-%s", testName, chainCfg.ChainID, RandLowerCaseLetterString(3))
		var err error
		users[i], err = chain.BuildWallet(ctx, keyName, "")
		if err != nil {
			return nil, fmt.Errorf("failed to get source user wallet: %w", err)
		}
	}

	return users, nil
}

func TestNonRefundable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	var (
		ctx                          = context.Background()
		client, network              = interchaintest.DockerSetup(t)
		rep                          = testreporter.NewNopReporter()
		eRep                         = rep.RelayerExecReporter(t)
		chainIdA, chainIdB, chainIdC = "chain-1", "chain-2", "chain-3"
		waitBlocks                   = 3
	)

	vals := 1
	fullNodes := 0

	baseCfg := DefaultConfig

	baseCfg.ChainID = chainIdA
	configA := baseCfg

	configB := NonRefundableConfig
	configB.ChainID = chainIdB

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
		ibc.CosmosRly,
		zaptest.NewLogger(t),
		relayer.DockerImage(&DefaultRelayer),
		relayer.StartupFlags("--processor", "events", "--block-history", "100"),
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
		TestName:  t.Name(),
		Client:    client,
		NetworkID: network,

		SkipPathCreation: false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	initBal := math.NewInt(10_000_000_000)

	users1 := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), initBal, chainA, chainC)
	users2 := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), initBal, chainA, chainC)
	users3 := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), initBal, chainA, chainC)
	users4 := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), initBal, chainA, chainC)
	users5 := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), initBal, chainA, chainC)
	users6 := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), initBal, chainA, chainC)
	users7 := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), initBal, chainA, chainC)
	users8 := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), initBal, chainA, chainC)
	usersA := []ibc.Wallet{users1[0], users2[0], users3[0], users4[0], users5[0], users6[0], users7[0], users8[0]}
	usersC := []ibc.Wallet{users1[1], users2[1], users3[1], users4[1], users5[1], users6[1], users7[1], users8[1]}

	usersB, err := BuildWallets(ctx, t.Name(), chainB, chainB, chainB, chainB)
	require.NoError(t, err)

	timeoutUsersC, err := BuildWallets(ctx, t.Name(), chainC, chainC, chainC, chainC)
	require.NoError(t, err)

	mintVoucherUsersA := usersA[4:]
	mintVoucherUsersC := usersC[4:]

	const invalidAddr = "xyz1t8eh66t2w5k67kwurmn5gqhtq6d2ja0vp7jmmq"

	abChan, err := ibc.GetTransferChannel(ctx, r, eRep, chainIdA, chainIdB)
	require.NoError(t, err)

	baChan := abChan.Counterparty

	cbChan, err := ibc.GetTransferChannel(ctx, r, eRep, chainIdC, chainIdB)
	require.NoError(t, err)

	bcChan := cbChan.Counterparty

	// Start the relayer on both paths
	err = r.StartRelayer(ctx, eRep, pathAB, pathBC)
	require.NoError(t, err)

	t.Cleanup(
		func() {
			err := r.StopRelayer(ctx, eRep)
			if err != nil {
				t.Logf("an error occured while stopping the relayer: %s", err)
			}
		},
	)

	// Compose the prefixed denoms and ibc denom for asserting balances
	firstHopDenom := transfertypes.GetPrefixedDenom(baChan.PortID, baChan.ChannelID, configA.Denom)
	secondHopDenom := transfertypes.GetPrefixedDenom(cbChan.PortID, cbChan.ChannelID, firstHopDenom)

	firstHopDenomTrace := transfertypes.ParseDenomTrace(firstHopDenom)
	secondHopDenomTrace := transfertypes.ParseDenomTrace(secondHopDenom)

	firstHopIBCDenom := firstHopDenomTrace.IBCDenom()
	secondHopIBCDenom := secondHopDenomTrace.IBCDenom()

	firstHopEscrowAccount := sdk.MustBech32ifyAddressBytes(configA.Bech32Prefix, transfertypes.GetEscrowAddress(abChan.PortID, abChan.ChannelID))
	secondHopEscrowAccount := sdk.MustBech32ifyAddressBytes(configB.Bech32Prefix, transfertypes.GetEscrowAddress(bcChan.PortID, bcChan.ChannelID))

	zeroBal := math.ZeroInt()
	transferAmount := math.NewInt(100_000)

	expectedFirstHopEscrowAmount := math.NewInt(0)

	t.Run("forward ack error refund - invalid receiver account on B", func(t *testing.T) {
		userA := usersA[0]

		// Send a malformed packet with invalid receiver address from Chain A->Chain B->Chain C
		transfer := ibc.WalletAmount{
			Address: "pfm",
			Denom:   configA.Denom,
			Amount:  transferAmount,
		}

		metadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: invalidAddr, // malformed receiver address on Chain C
				Channel:  bcChan.ChannelID,
				Port:     bcChan.PortID,
			},
		}

		memo, err := json.Marshal(metadata)
		require.NoError(t, err)

		chainAHeight, err := chainA.Height(ctx)
		require.NoError(t, err)

		transferTx, err := chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName(), transfer, ibc.TransferOptions{Memo: string(memo)})
		require.NoError(t, err)

		_, err = testutil.PollForAck(ctx, chainA, chainAHeight, chainAHeight+25, transferTx.Packet)
		require.NoError(t, err)
		err = testutil.WaitForBlocks(ctx, waitBlocks, chainA)
		require.NoError(t, err)

		// assert balances for user controlled wallets
		chainABalance, err := chainA.GetBalance(ctx, userA.FormattedAddress(), configA.Denom)
		require.NoError(t, err)

		// funds should end up in the A user's bech32 transformed address on chain B.
		chainBBalance, err := chainB.GetBalance(ctx, userA.FormattedAddress(), firstHopIBCDenom)
		require.NoError(t, err)

		require.True(t, chainABalance.Equal(initBal.Sub(transferAmount)))

		// funds should end up in the A user's bech32 transformed address on chain B.
		require.True(t, chainBBalance.Equal(transferAmount))

		// assert balances for IBC escrow accounts
		firstHopEscrowBalance, err := chainA.GetBalance(ctx, firstHopEscrowAccount, chainA.Config().Denom)
		require.NoError(t, err)

		secondHopEscrowBalance, err := chainB.GetBalance(ctx, secondHopEscrowAccount, firstHopIBCDenom)
		require.NoError(t, err)

		expectedFirstHopEscrowAmount = expectedFirstHopEscrowAmount.Add(transferAmount)

		require.True(t, firstHopEscrowBalance.Equal(expectedFirstHopEscrowAmount))
		require.True(t, secondHopEscrowBalance.Equal(zeroBal))
	})

	t.Run("forward ack error refund - valid receiver account on B", func(t *testing.T) {
		userA := usersA[1]
		userB := usersB[0]
		// Send a malformed packet with valid receiver address from Chain A->Chain B->Chain C
		transfer := ibc.WalletAmount{
			Address: userB.FormattedAddress(),
			Denom:   configA.Denom,
			Amount:  transferAmount,
		}

		metadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: invalidAddr, // malformed receiver address on Chain C
				Channel:  bcChan.ChannelID,
				Port:     bcChan.PortID,
			},
		}

		memo, err := json.Marshal(metadata)
		require.NoError(t, err)

		chainAHeight, err := chainA.Height(ctx)
		require.NoError(t, err)

		transferTx, err := chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName(), transfer, ibc.TransferOptions{Memo: string(memo)})
		require.NoError(t, err)

		_, err = testutil.PollForAck(ctx, chainA, chainAHeight, chainAHeight+25, transferTx.Packet)
		require.NoError(t, err)
		err = testutil.WaitForBlocks(ctx, waitBlocks, chainA)
		require.NoError(t, err)

		// assert balances for user controlled wallets
		chainABalance, err := chainA.GetBalance(ctx, userA.FormattedAddress(), chainA.Config().Denom)
		require.NoError(t, err)

		chainBBalance, err := chainB.GetBalance(ctx, userB.FormattedAddress(), firstHopIBCDenom)
		require.NoError(t, err)

		require.True(t, chainABalance.Equal(initBal.Sub(transferAmount)))
		require.True(t, chainBBalance.Equal(transferAmount))

		// assert balances for IBC escrow accounts
		firstHopEscrowBalance, err := chainA.GetBalance(ctx, firstHopEscrowAccount, chainA.Config().Denom)
		require.NoError(t, err)

		secondHopEscrowBalance, err := chainB.GetBalance(ctx, secondHopEscrowAccount, firstHopIBCDenom)
		require.NoError(t, err)

		expectedFirstHopEscrowAmount = expectedFirstHopEscrowAmount.Add(transferAmount)

		require.True(t, firstHopEscrowBalance.Equal(expectedFirstHopEscrowAmount))
		require.True(t, secondHopEscrowBalance.Equal(zeroBal))
	})

	t.Run("forward timeout refund - valid receiver account on B", func(t *testing.T) {
		userA := usersA[2]
		userB := usersB[1]
		userC := timeoutUsersC[0]
		// Send packet from Chain A->Chain B->Chain C with the timeout so low for B->C transfer that it can not make it from B to C,
		// which should result in a refund to User B on Chain B after two retries.
		transfer := ibc.WalletAmount{
			Address: userB.FormattedAddress(),
			Denom:   configA.Denom,
			Amount:  transferAmount,
		}

		retries := uint8(2)
		metadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: userC.FormattedAddress(),
				Channel:  bcChan.ChannelID,
				Port:     bcChan.PortID,
				Retries:  &retries,
				Timeout:  1 * time.Second,
			},
		}

		memo, err := json.Marshal(metadata)
		require.NoError(t, err)

		chainAHeight, err := chainA.Height(ctx)
		require.NoError(t, err)

		transferTx, err := chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName(), transfer, ibc.TransferOptions{Memo: string(memo)})
		require.NoError(t, err)
		_, err = testutil.PollForAck(ctx, chainA, chainAHeight, chainAHeight+25, transferTx.Packet)
		require.NoError(t, err)
		err = testutil.WaitForBlocks(ctx, waitBlocks, chainA)
		require.NoError(t, err)

		// assert balances for user controlled wallets
		chainABalance, err := chainA.GetBalance(ctx, userA.FormattedAddress(), chainA.Config().Denom)
		require.NoError(t, err)

		chainBBalance, err := chainB.GetBalance(ctx, userB.FormattedAddress(), firstHopIBCDenom)
		require.NoError(t, err)

		chainCBalance, err := chainC.GetBalance(ctx, userC.FormattedAddress(), secondHopIBCDenom)
		require.NoError(t, err)

		require.True(t, chainABalance.Equal(initBal.Sub(transferAmount)))
		require.True(t, chainBBalance.Equal(transferAmount))
		require.True(t, chainCBalance.Equal(zeroBal))

		firstHopEscrowBalance, err := chainA.GetBalance(ctx, firstHopEscrowAccount, chainA.Config().Denom)
		require.NoError(t, err)

		secondHopEscrowBalance, err := chainB.GetBalance(ctx, secondHopEscrowAccount, firstHopIBCDenom)
		require.NoError(t, err)

		expectedFirstHopEscrowAmount = expectedFirstHopEscrowAmount.Add(transferAmount)

		require.True(t, firstHopEscrowBalance.Equal(expectedFirstHopEscrowAmount))
		require.True(t, secondHopEscrowBalance.Equal(zeroBal))
	})

	t.Run("forward timeout refund - invalid receiver account on B", func(t *testing.T) {
		userA := usersA[3]
		userC := timeoutUsersC[1]
		// Send packet from Chain A->Chain B->Chain C with the timeout so low for B->C transfer that it can not make it from B to C,
		// which should result in a refund to the bech32 equivalent of userA on chain B after two retries.
		transfer := ibc.WalletAmount{
			Address: "pfm",
			Denom:   configA.Denom,
			Amount:  transferAmount,
		}

		retries := uint8(2)
		metadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: userC.FormattedAddress(),
				Channel:  bcChan.ChannelID,
				Port:     bcChan.PortID,
				Retries:  &retries,
				Timeout:  1 * time.Second,
			},
		}

		memo, err := json.Marshal(metadata)
		require.NoError(t, err)

		chainAHeight, err := chainA.Height(ctx)
		require.NoError(t, err)

		transferTx, err := chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName(), transfer, ibc.TransferOptions{Memo: string(memo)})
		require.NoError(t, err)
		_, err = testutil.PollForAck(ctx, chainA, chainAHeight, chainAHeight+25, transferTx.Packet)
		require.NoError(t, err)
		err = testutil.WaitForBlocks(ctx, waitBlocks, chainA)
		require.NoError(t, err)

		// assert balances for user controlled wallets
		chainABalance, err := chainA.GetBalance(ctx, userA.FormattedAddress(), chainA.Config().Denom)
		require.NoError(t, err)

		chainBBalance, err := chainB.GetBalance(ctx, userA.FormattedAddress(), firstHopIBCDenom)
		require.NoError(t, err)

		chainCBalance, err := chainC.GetBalance(ctx, userC.FormattedAddress(), secondHopIBCDenom)
		require.NoError(t, err)

		require.True(t, chainABalance.Equal(initBal.Sub(transferAmount)))
		require.True(t, chainBBalance.Equal(transferAmount))
		require.True(t, chainCBalance.Equal(zeroBal))

		firstHopEscrowBalance, err := chainA.GetBalance(ctx, firstHopEscrowAccount, chainA.Config().Denom)
		require.NoError(t, err)

		secondHopEscrowBalance, err := chainB.GetBalance(ctx, secondHopEscrowAccount, firstHopIBCDenom)
		require.NoError(t, err)

		expectedFirstHopEscrowAmount = expectedFirstHopEscrowAmount.Add(transferAmount)

		require.True(t, firstHopEscrowBalance.Equal(expectedFirstHopEscrowAmount))
		require.True(t, secondHopEscrowBalance.Equal(zeroBal))
	})

	revFirstHopDenom := transfertypes.GetPrefixedDenom(bcChan.PortID, bcChan.ChannelID, configC.Denom)
	revSecondHopDenom := transfertypes.GetPrefixedDenom(abChan.PortID, abChan.ChannelID, revFirstHopDenom)

	revFirstHopDenomTrace := transfertypes.ParseDenomTrace(revFirstHopDenom)
	revSecondHopDenomTrace := transfertypes.ParseDenomTrace(revSecondHopDenom)

	revFirstHopIBCDenom := revFirstHopDenomTrace.IBCDenom()
	revSecondHopIBCDenom := revSecondHopDenomTrace.IBCDenom()

	revFirstHopEscrowAccount := sdk.MustBech32ifyAddressBytes(configC.Bech32Prefix, transfertypes.GetEscrowAddress(cbChan.PortID, cbChan.ChannelID))
	revSecondHopEscrowAccount := sdk.MustBech32ifyAddressBytes(configB.Bech32Prefix, transfertypes.GetEscrowAddress(baChan.PortID, baChan.ChannelID))

	expectedRevFirstHopEscrowAmount := math.NewInt(0)
	expectedRevSecondHopEscrowAmount := math.NewInt(0)

	var eg errgroup.Group
	for i, userC := range mintVoucherUsersC {
		userC := userC
		userA := mintVoucherUsersA[i]
		eg.Go(func() error {
			// Send packet from Chain C->Chain B->Chain A
			transfer := ibc.WalletAmount{
				Address: "pfm",
				Denom:   configC.Denom,
				Amount:  transferAmount,
			}

			forwardMetadata := &PacketMetadata{
				Forward: &ForwardMetadata{
					Receiver: userA.FormattedAddress(),
					Channel:  baChan.ChannelID,
					Port:     baChan.PortID,
				},
			}

			memo, err := json.Marshal(forwardMetadata)
			if err != nil {
				return err
			}

			chainCHeight, err := chainC.Height(ctx)
			if err != nil {
				return err
			}

			transferTx, err := chainC.SendIBCTransfer(ctx, cbChan.ChannelID, userC.KeyName(), transfer, ibc.TransferOptions{Memo: string(memo)})
			if err != nil {
				return err
			}
			_, err = testutil.PollForAck(ctx, chainC, chainCHeight, chainCHeight+30, transferTx.Packet)
			if err != nil {
				return err
			}
			err = testutil.WaitForBlocks(ctx, waitBlocks, chainC)
			if err != nil {
				return err
			}

			chainCBalance, err := chainC.GetBalance(ctx, userC.FormattedAddress(), configC.Denom)
			if err != nil {
				return err
			}

			// TODO add balance check for intermediate account on B

			chainABalance, err := chainA.GetBalance(ctx, userA.FormattedAddress(), revSecondHopIBCDenom)
			if err != nil {
				return err
			}

			chainCExpected := initBal.Sub(transferAmount)

			if !chainCBalance.Equal(chainCExpected) {
				return fmt.Errorf("chain c expected balance %s, got %s", chainCExpected, chainCBalance)
			}

			if !chainABalance.Equal(transferAmount) {
				return fmt.Errorf("chain a expected balance %s, got %s", transferAmount, chainABalance)
			}

			return nil
		})
	}

	require.NoError(t, eg.Wait())

	expectedRevFirstHopEscrowAmount = expectedRevFirstHopEscrowAmount.Add(transferAmount.Mul(math.NewIntFromUint64(uint64(len(mintVoucherUsersC)))))
	expectedRevSecondHopEscrowAmount = expectedRevSecondHopEscrowAmount.Add(transferAmount.Mul(math.NewIntFromUint64(uint64(len(mintVoucherUsersC)))))

	// assert balances for IBC escrow accounts
	revFirstHopEscrowBalance, err := chainC.GetBalance(ctx, revFirstHopEscrowAccount, configC.Denom)
	require.NoError(t, err)

	revSecondHopEscrowBalance, err := chainB.GetBalance(ctx, revSecondHopEscrowAccount, revFirstHopIBCDenom)
	require.NoError(t, err)

	require.True(t, revFirstHopEscrowBalance.Equal(expectedRevFirstHopEscrowAmount))
	require.True(t, revSecondHopEscrowBalance.Equal(expectedRevSecondHopEscrowAmount))

	t.Run("rev forward ack error refund - invalid receiver account on B", func(t *testing.T) {
		userA := mintVoucherUsersA[0]

		// Send a malformed packet with invalid receiver address from Chain A->Chain B->Chain C
		transfer := ibc.WalletAmount{
			Address: "pfm",
			Denom:   revSecondHopIBCDenom,
			Amount:  transferAmount,
		}

		metadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: invalidAddr, // malformed receiver address on Chain C
				Channel:  bcChan.ChannelID,
				Port:     bcChan.PortID,
			},
		}

		memo, err := json.Marshal(metadata)
		require.NoError(t, err)

		chainAHeight, err := chainA.Height(ctx)
		require.NoError(t, err)

		transferTx, err := chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName(), transfer, ibc.TransferOptions{Memo: string(memo)})
		require.NoError(t, err)

		_, err = testutil.PollForAck(ctx, chainA, chainAHeight, chainAHeight+25, transferTx.Packet)
		require.NoError(t, err)
		err = testutil.WaitForBlocks(ctx, waitBlocks, chainA)
		require.NoError(t, err)

		// assert balances for user controlled wallets
		chainABalance, err := chainA.GetBalance(ctx, userA.FormattedAddress(), revSecondHopIBCDenom)
		require.NoError(t, err)

		// funds should end up in the A user's bech32 transformed address on chain B.
		chainBBalance, err := chainB.GetBalance(ctx, userA.FormattedAddress(), revFirstHopIBCDenom)
		require.NoError(t, err)

		require.True(t, chainABalance.Equal(zeroBal))

		// funds should end up in the A user's bech32 transformed address on chain B.
		require.True(t, chainBBalance.Equal(transferAmount))

		// assert balances for IBC escrow accounts
		revFirstHopEscrowBalance, err := chainC.GetBalance(ctx, revFirstHopEscrowAccount, configC.Denom)
		require.NoError(t, err)

		revSecondHopEscrowBalance, err := chainB.GetBalance(ctx, revSecondHopEscrowAccount, revFirstHopIBCDenom)
		require.NoError(t, err)

		expectedRevSecondHopEscrowAmount = expectedRevSecondHopEscrowAmount.Sub(transferAmount)

		require.Truef(t, revFirstHopEscrowBalance.Equal(expectedRevFirstHopEscrowAmount), "expected %s, got %s", expectedRevFirstHopEscrowAmount, revFirstHopEscrowBalance)
		require.Truef(t, revSecondHopEscrowBalance.Equal(expectedRevSecondHopEscrowAmount), "expected %s, got %s", expectedRevSecondHopEscrowAmount, revSecondHopEscrowBalance)
	})

	t.Run("rev forward ack error refund - valid receiver account on B", func(t *testing.T) {
		userA := mintVoucherUsersA[1]
		userB := usersB[2]
		// Send a malformed packet with valid receiver address from Chain A->Chain B->Chain C
		transfer := ibc.WalletAmount{
			Address: userB.FormattedAddress(),
			Denom:   revSecondHopIBCDenom,
			Amount:  transferAmount,
		}

		metadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: invalidAddr, // malformed receiver address on Chain C
				Channel:  bcChan.ChannelID,
				Port:     bcChan.PortID,
			},
		}

		memo, err := json.Marshal(metadata)
		require.NoError(t, err)

		chainAHeight, err := chainA.Height(ctx)
		require.NoError(t, err)

		transferTx, err := chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName(), transfer, ibc.TransferOptions{Memo: string(memo)})
		require.NoError(t, err)

		_, err = testutil.PollForAck(ctx, chainA, chainAHeight, chainAHeight+25, transferTx.Packet)
		require.NoError(t, err)
		err = testutil.WaitForBlocks(ctx, waitBlocks, chainA)
		require.NoError(t, err)

		// assert balances for user controlled wallets
		chainABalance, err := chainA.GetBalance(ctx, userA.FormattedAddress(), revSecondHopIBCDenom)
		require.NoError(t, err)

		chainBBalance, err := chainB.GetBalance(ctx, userB.FormattedAddress(), revFirstHopIBCDenom)
		require.NoError(t, err)

		require.True(t, chainABalance.Equal(zeroBal))
		require.True(t, chainBBalance.Equal(transferAmount))

		// assert balances for IBC escrow accounts
		revFirstHopEscrowBalance, err := chainC.GetBalance(ctx, revFirstHopEscrowAccount, configC.Denom)
		require.NoError(t, err)

		revSecondHopEscrowBalance, err := chainB.GetBalance(ctx, revSecondHopEscrowAccount, revFirstHopIBCDenom)
		require.NoError(t, err)

		expectedRevSecondHopEscrowAmount = expectedRevSecondHopEscrowAmount.Sub(transferAmount)

		require.Truef(t, revFirstHopEscrowBalance.Equal(expectedRevFirstHopEscrowAmount), "expected %s, got %s", expectedRevFirstHopEscrowAmount, revFirstHopEscrowBalance)
		require.Truef(t, revSecondHopEscrowBalance.Equal(expectedRevSecondHopEscrowAmount), "expected %s, got %s", expectedRevSecondHopEscrowAmount, revSecondHopEscrowBalance)
	})

	t.Run("rev forward timeout refund - valid receiver account on B", func(t *testing.T) {
		userA := mintVoucherUsersA[2]
		userB := usersB[3]
		userC := timeoutUsersC[2]
		// Send packet from Chain A->Chain B->Chain C with the timeout so low for B->C transfer that it can not make it from B to C,
		// which should result in a refund to User B on Chain B after two retries.
		transfer := ibc.WalletAmount{
			Address: userB.FormattedAddress(),
			Denom:   revSecondHopIBCDenom,
			Amount:  transferAmount,
		}

		retries := uint8(2)
		metadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: userC.FormattedAddress(),
				Channel:  bcChan.ChannelID,
				Port:     bcChan.PortID,
				Retries:  &retries,
				Timeout:  1 * time.Second,
			},
		}

		memo, err := json.Marshal(metadata)
		require.NoError(t, err)

		chainAHeight, err := chainA.Height(ctx)
		require.NoError(t, err)

		transferTx, err := chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName(), transfer, ibc.TransferOptions{Memo: string(memo)})
		require.NoError(t, err)
		_, err = testutil.PollForAck(ctx, chainA, chainAHeight, chainAHeight+25, transferTx.Packet)
		require.NoError(t, err)
		err = testutil.WaitForBlocks(ctx, waitBlocks, chainA)
		require.NoError(t, err)

		// assert balances for user controlled wallets
		chainABalance, err := chainA.GetBalance(ctx, userA.FormattedAddress(), revSecondHopIBCDenom)
		require.NoError(t, err)

		chainBBalance, err := chainB.GetBalance(ctx, userB.FormattedAddress(), revFirstHopIBCDenom)
		require.NoError(t, err)

		chainCBalance, err := chainC.GetBalance(ctx, userC.FormattedAddress(), configC.Denom)
		require.NoError(t, err)

		require.Truef(t, chainABalance.Equal(zeroBal), "chain a balance, expected %s, got %s", zeroBal, chainABalance)
		require.Truef(t, chainBBalance.Equal(transferAmount), "chain b balance, expected %s, got %s", transferAmount, chainBBalance)
		require.Truef(t, chainCBalance.Equal(zeroBal), "chain c balance, expected %s, got %s", zeroBal, chainCBalance)

		revFirstHopEscrowBalance, err := chainC.GetBalance(ctx, revFirstHopEscrowAccount, configC.Denom)
		require.NoError(t, err)

		revSecondHopEscrowBalance, err := chainB.GetBalance(ctx, revSecondHopEscrowAccount, revFirstHopIBCDenom)
		require.NoError(t, err)

		expectedRevSecondHopEscrowAmount = expectedRevSecondHopEscrowAmount.Sub(transferAmount)

		require.Truef(t, revFirstHopEscrowBalance.Equal(expectedRevFirstHopEscrowAmount), "expected %s, got %s", expectedRevFirstHopEscrowAmount, revFirstHopEscrowBalance)
		require.Truef(t, revSecondHopEscrowBalance.Equal(expectedRevSecondHopEscrowAmount), "expected %s, got %s", expectedRevSecondHopEscrowAmount, revSecondHopEscrowBalance)
	})

	t.Run("rev forward timeout refund - invalid receiver account on B", func(t *testing.T) {
		userA := mintVoucherUsersA[3]
		userC := timeoutUsersC[3]
		// Send packet from Chain A->Chain B->Chain C with the timeout so low for B->C transfer that it can not make it from B to C,
		// which should result in a refund to the bech32 equivalent of userA on chain B after two retries.
		transfer := ibc.WalletAmount{
			Address: "pfm",
			Denom:   revSecondHopIBCDenom,
			Amount:  transferAmount,
		}

		retries := uint8(2)
		metadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: userC.FormattedAddress(),
				Channel:  bcChan.ChannelID,
				Port:     bcChan.PortID,
				Retries:  &retries,
				Timeout:  1 * time.Second,
			},
		}

		memo, err := json.Marshal(metadata)
		require.NoError(t, err)

		chainAHeight, err := chainA.Height(ctx)
		require.NoError(t, err)

		transferTx, err := chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName(), transfer, ibc.TransferOptions{Memo: string(memo)})
		require.NoError(t, err)
		_, err = testutil.PollForAck(ctx, chainA, chainAHeight, chainAHeight+25, transferTx.Packet)
		require.NoError(t, err)
		err = testutil.WaitForBlocks(ctx, waitBlocks, chainA)
		require.NoError(t, err)

		// assert balances for user controlled wallets
		chainABalance, err := chainA.GetBalance(ctx, userA.FormattedAddress(), revSecondHopIBCDenom)
		require.NoError(t, err)

		chainBBalance, err := chainB.GetBalance(ctx, userA.FormattedAddress(), revFirstHopIBCDenom)
		require.NoError(t, err)

		chainCBalance, err := chainC.GetBalance(ctx, userC.FormattedAddress(), configC.Denom)
		require.NoError(t, err)

		require.Truef(t, chainABalance.Equal(zeroBal), "chain a balance, expected %s, got %s", zeroBal, chainABalance)
		require.Truef(t, chainBBalance.Equal(transferAmount), "chain b balance, expected %s, got %s", transferAmount, chainBBalance)
		require.Truef(t, chainCBalance.Equal(zeroBal), "chain c balance, expected %s, got %s", zeroBal, chainCBalance)

		revFirstHopEscrowBalance, err := chainC.GetBalance(ctx, revFirstHopEscrowAccount, configC.Denom)
		require.NoError(t, err)

		revSecondHopEscrowBalance, err := chainB.GetBalance(ctx, revSecondHopEscrowAccount, revFirstHopIBCDenom)
		require.NoError(t, err)

		expectedRevSecondHopEscrowAmount = expectedRevSecondHopEscrowAmount.Sub(transferAmount)

		require.Truef(t, revFirstHopEscrowBalance.Equal(expectedRevFirstHopEscrowAmount), "expected %s, got %s", expectedRevFirstHopEscrowAmount, revFirstHopEscrowBalance)
		require.Truef(t, revSecondHopEscrowBalance.Equal(expectedRevSecondHopEscrowAmount), "expected %s, got %s", expectedRevSecondHopEscrowAmount, revSecondHopEscrowBalance)
	})
}
