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
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/ibc"
	icrelayer "github.com/cosmos/interchaintest/v10/relayer"
	"github.com/cosmos/interchaintest/v10/testreporter"
	"github.com/cosmos/interchaintest/v10/testutil"
	dockerclient "github.com/docker/docker/client"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zaptest"
)

var (
	chars                            = []byte("abcdefghijklmnopqrstuvwxyz")
	waitBlocks                       = 20
	chainIdA, chainIdB, chainIdC     = "chain-1", "chain-2", "chain-3"
	configA                          ibc.ChainConfig
	configB                          ibc.ChainConfig
	configC                          ibc.ChainConfig
	ctx                              context.Context
	dockerClient                     *dockerclient.Client
	network                          string
	rep                              = testreporter.NewNopReporter()
	eRep                             *testreporter.RelayerExecReporter
	firstHopIBCDenom                 string
	secondHopIBCDenom                string
	firstHopEscrowAccount            string
	secondHopEscrowAccount           string
	zeroBal                          = math.ZeroInt()
	initBal                          = math.NewInt(10_000_000_000)
	transferAmount                   = math.NewInt(100_000)
	invalidAddr                      = "xyz1t8eh66t2w5k67kwurmn5gqhtq6d2ja0vp7jmmq"
	expectedFirstHopEscrowAmount     = math.NewInt(0)
	chainA                           *cosmos.CosmosChain
	chainB                           *cosmos.CosmosChain
	chainC                           *cosmos.CosmosChain
	ibcRelayer                       ibc.Relayer
	pathAB                           = "ab"
	pathBC                           = "bc"
	abChan                           *ibc.ChannelOutput
	baChan                           ibc.ChannelCounterparty
	cbChan                           *ibc.ChannelOutput
	bcChan                           ibc.ChannelCounterparty
	revFirstHopIBCDenom              string
	revSecondHopIBCDenom             string
	revFirstHopEscrowAccount         string
	revSecondHopEscrowAccount        string
	expectedRevFirstHopEscrowAmount  = math.NewInt(0)
	expectedRevSecondHopEscrowAmount = math.NewInt(0)
)

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
		t.Skip()
	}
	suite.Run(t, new(TestPFMSuite))
}

type TestPFMSuite struct {
	suite.Suite
}

func (s *TestPFMSuite) SetupSuite() {
	require := s.Require()
	ctx = context.Background()
	dockerClient, network = interchaintest.DockerSetup(s.T())
	eRep = rep.RelayerExecReporter(s.T())

	vals := 1
	fullNodes := 0

	baseCfg := DefaultConfig

	baseCfg.ChainID = chainIdA
	cfgA := baseCfg
	configA = cfgA

	cfgB := NonRefundableConfig
	cfgB.ChainID = chainIdB
	configB = cfgB

	cfgC := baseCfg
	cfgC.ChainID = chainIdC
	configC = cfgC

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(s.T()), []*interchaintest.ChainSpec{
		{Name: "pfm", ChainConfig: configA, NumFullNodes: &fullNodes, NumValidators: &vals},
		{Name: "pfm", ChainConfig: configB, NumFullNodes: &fullNodes, NumValidators: &vals},
		{Name: "pfm", ChainConfig: configC, NumFullNodes: &fullNodes, NumValidators: &vals},
	})

	chains, err := cf.Chains(s.T().Name())
	require.NoError(err)

	chainA, chainB, chainC = chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain), chains[2].(*cosmos.CosmosChain)

	ibcRelayer = interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(s.T()),
		icrelayer.DockerImage(&DefaultRelayer),
	).Build(s.T(), dockerClient, network)

	ic := interchaintest.NewInterchain().
		AddChain(chainA).
		AddChain(chainB).
		AddChain(chainC).
		AddRelayer(ibcRelayer, "hermes").
		AddLink(interchaintest.InterchainLink{
			Chain1:  chainA,
			Chain2:  chainB,
			Relayer: ibcRelayer,
			Path:    pathAB,
		}).
		AddLink(interchaintest.InterchainLink{
			Chain1:  chainB,
			Chain2:  chainC,
			Relayer: ibcRelayer,
			Path:    pathBC,
		})

	require.NoError(ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:  s.T().Name(),
		Client:    dockerClient,
		NetworkID: network,

		SkipPathCreation: false,
	}))

	s.T().Cleanup(func() {
		_ = ic.Close()
	})

	abChan, err = ibc.GetTransferChannel(ctx, ibcRelayer, eRep, chainIdA, chainIdB)
	require.NoError(err)
	baChan = abChan.Counterparty
	cbChan, err = ibc.GetTransferChannel(ctx, ibcRelayer, eRep, chainIdC, chainIdB)
	require.NoError(err)
	bcChan = cbChan.Counterparty

	// Start the relayer on both paths
	err = ibcRelayer.StartRelayer(ctx, eRep, pathAB, pathBC)
	require.NoError(err)

	s.T().Cleanup(func() {
		err := ibcRelayer.StopRelayer(ctx, eRep)
		if err != nil {
			s.T().Logf("an error occured while stopping the relayer: %s", err)
		}
	})

	// Compose the prefixed denoms and ibc denom for asserting balances
	firstHopDenom := transfertypes.NewDenom(configA.Denom, transfertypes.Hop{
		PortId:    baChan.PortID,
		ChannelId: baChan.ChannelID,
	})
	secondHopDenom := transfertypes.NewDenom(firstHopDenom.Base, transfertypes.Hop{
		PortId:    cbChan.PortID,
		ChannelId: cbChan.ChannelID,
	})
	firstHopIBCDenom = firstHopDenom.IBCDenom()
	secondHopIBCDenom = secondHopDenom.IBCDenom()

	firstHopEscrowAccount = sdk.MustBech32ifyAddressBytes(configA.Bech32Prefix, transfertypes.GetEscrowAddress(abChan.PortID, abChan.ChannelID))
	secondHopEscrowAccount = sdk.MustBech32ifyAddressBytes(configB.Bech32Prefix, transfertypes.GetEscrowAddress(bcChan.PortID, bcChan.ChannelID))

	revFirstHopDenom := transfertypes.NewDenom(configC.Denom, transfertypes.Hop{
		PortId:    bcChan.PortID,
		ChannelId: bcChan.ChannelID,
	})
	revSecondHopDenom := transfertypes.NewDenom(revFirstHopDenom.Path(), transfertypes.Hop{
		PortId:    abChan.PortID,
		ChannelId: abChan.ChannelID,
	})

	revFirstHopIBCDenom = revFirstHopDenom.IBCDenom()
	revSecondHopIBCDenom = revSecondHopDenom.IBCDenom()

	revFirstHopEscrowAccount = sdk.MustBech32ifyAddressBytes(configC.Bech32Prefix, transfertypes.GetEscrowAddress(cbChan.PortID, cbChan.ChannelID))
	revSecondHopEscrowAccount = sdk.MustBech32ifyAddressBytes(configB.Bech32Prefix, transfertypes.GetEscrowAddress(baChan.PortID, baChan.ChannelID))
}

func (s *TestPFMSuite) TestForwardAckErrorInvalidReceiver() {
	require := s.Require()
	users := interchaintest.GetAndFundTestUsers(s.T(), ctx, s.T().Name(), initBal, chainA)
	userA := users[0]

	chainABalanceBefore, err := chainA.GetBalance(ctx, userA.FormattedAddress(), configA.Denom)
	require.NoError(err)
	fmt.Println("UserA balance on chain A before", chainABalanceBefore.String(), userA.FormattedAddress())
	chainBBalanceBefore, err := chainB.GetBalance(ctx, userA.FormattedAddress(), firstHopIBCDenom)
	require.NoError(err)
	fmt.Println("UserA balance on chain B before", chainBBalanceBefore.String(), userA.FormattedAddress())
	firstHopEscrowBalanceBefore, err := chainA.GetBalance(ctx, firstHopEscrowAccount, chainA.Config().Denom)
	require.NoError(err)
	fmt.Println("First hop escrow balance on chain A before", firstHopEscrowBalanceBefore.String(), firstHopEscrowAccount)

	secondHopEscrowBalanceBefore, err := chainB.GetBalance(ctx, secondHopEscrowAccount, firstHopIBCDenom)
	require.NoError(err)
	fmt.Println("Second hop escrow balance on chain B before", secondHopEscrowBalanceBefore.String(), secondHopEscrowAccount)

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
	require.NoError(err)

	_, err = chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName(), transfer,
		ibc.TransferOptions{Memo: string(memo)})
	require.NoError(err)

	err = testutil.WaitForBlocks(ctx, waitBlocks, chainA)
	require.NoError(err)
	err = ibcRelayer.Flush(ctx, eRep, pathAB, abChan.ChannelID)
	require.NoError(err)
	err = ibcRelayer.Flush(ctx, eRep, pathBC, bcChan.ChannelID)
	require.NoError(err)

	// assert balances for user controlled wallets
	chainABalance, err := chainA.GetBalance(ctx, userA.FormattedAddress(), configA.Denom)
	require.NoError(err)
	fmt.Println("UserA balance on chain A before", chainABalance.String(), userA.FormattedAddress())

	// funds should end up in the A user's bech32 transformed address on chain B.
	chainBBalance, err := chainB.GetBalance(ctx, userA.FormattedAddress(), firstHopIBCDenom)
	require.NoError(err)
	fmt.Println("UserA balance on chain B after", chainBBalance.String(), userA.FormattedAddress())

	chainCBalance, err := chainC.GetBalance(ctx, userA.FormattedAddress(), secondHopIBCDenom)
	require.NoError(err)
	fmt.Println("UserA balance on chain C after", chainCBalance.String(), userA.FormattedAddress())

	require.True(chainABalance.Equal(initBal.Sub(transferAmount)))

	// funds should end up in the A user's bech32 transformed address on chain B.
	require.True(chainBBalance.Equal(transferAmount))

	// assert balances for IBC escrow accounts
	firstHopEscrowBalance, err := chainA.GetBalance(ctx, firstHopEscrowAccount, chainA.Config().Denom)
	require.NoError(err)
	fmt.Println("First hop escrow balance on chain A after", firstHopEscrowBalance.String(), firstHopEscrowAccount)

	secondHopEscrowBalance, err := chainB.GetBalance(ctx, secondHopEscrowAccount, firstHopIBCDenom)
	require.NoError(err)
	fmt.Println("Second hop escrow balance on chain B after", secondHopEscrowBalance.String(), secondHopEscrowAccount)

	expectedFirstHopEscrowAmount = expectedFirstHopEscrowAmount.Add(transferAmount)

	require.True(firstHopEscrowBalance.Equal(expectedFirstHopEscrowAmount))
	require.True(secondHopEscrowBalance.Equal(zeroBal))
}

func (s *TestPFMSuite) TestForwardAckErrorValidReceiver() {
	require := s.Require()
	users := interchaintest.GetAndFundTestUsers(s.T(), ctx, s.T().Name(), initBal, chainA, chainB)
	userA := users[0]
	userB := users[1]
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
	require.NoError(err)

	_, err = chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName(), transfer,
		ibc.TransferOptions{Memo: string(memo)})
	require.NoError(err)

	err = testutil.WaitForBlocks(ctx, waitBlocks, chainA)
	require.NoError(err)
	err = ibcRelayer.Flush(ctx, eRep, pathAB, abChan.ChannelID)
	require.NoError(err)

	// assert balances for user controlled wallets
	chainABalance, err := chainA.GetBalance(ctx, userA.FormattedAddress(), chainA.Config().Denom)
	require.NoError(err)

	chainBBalance, err := chainB.GetBalance(ctx, userB.FormattedAddress(), firstHopIBCDenom)
	require.NoError(err)

	require.True(chainABalance.Equal(initBal.Sub(transferAmount)))
	require.True(chainBBalance.Equal(transferAmount))

	// assert balances for IBC escrow accounts
	firstHopEscrowBalance, err := chainA.GetBalance(ctx, firstHopEscrowAccount, chainA.Config().Denom)
	require.NoError(err)

	secondHopEscrowBalance, err := chainB.GetBalance(ctx, secondHopEscrowAccount, firstHopIBCDenom)
	require.NoError(err)

	expectedFirstHopEscrowAmount = expectedFirstHopEscrowAmount.Add(transferAmount)

	require.True(firstHopEscrowBalance.Equal(expectedFirstHopEscrowAmount))
	require.True(secondHopEscrowBalance.Equal(zeroBal))
}

func (s *TestPFMSuite) TestTimeoutRefundValidReceiver() {
	require := s.Require()
	users := interchaintest.GetAndFundTestUsers(s.T(), ctx, s.T().Name(), initBal, chainA, chainB)
	userA := users[0]
	userB := users[1]
	wallets, err := BuildWallets(ctx, s.T().Name(), chainC)
	require.NoError(err)
	timeoutUser := wallets[0]

	chainABalanceBefore, err := chainA.GetBalance(ctx, userA.FormattedAddress(), chainA.Config().Denom)
	require.NoError(err)

	chainBBalanceBefore, err := chainB.GetBalance(ctx, userB.FormattedAddress(), firstHopIBCDenom)
	require.NoError(err)

	chainCBalanceBefore, err := chainC.GetBalance(ctx, timeoutUser.FormattedAddress(), secondHopIBCDenom)
	require.NoError(err)

	firstHopEscrowBalanceBefore, err := chainA.GetBalance(ctx, firstHopEscrowAccount, chainA.Config().Denom)
	require.NoError(err)

	secondHopEscrowBalanceBefore, err := chainB.GetBalance(ctx, secondHopEscrowAccount, firstHopIBCDenom)
	require.NoError(err)

	fmt.Printf("UserA balance on chain A before: %s (%s)\n", chainABalanceBefore.String(), userA.FormattedAddress())
	fmt.Printf("UserB balance on chain B before: %s (%s)\n", chainBBalanceBefore.String(), userB.FormattedAddress())
	fmt.Printf("UserC balance on chain C before: %s (%s)\n", chainCBalanceBefore.String(), timeoutUser.FormattedAddress())

	fmt.Printf("First hop escrow balance on chain A before: %s (%s)\n", firstHopEscrowBalanceBefore.String(), firstHopEscrowAccount)
	fmt.Printf("Second hop escrow balance on chain B before: %s (%s)\n", secondHopEscrowBalanceBefore.String(), secondHopEscrowAccount)

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
			Receiver: timeoutUser.FormattedAddress(),
			Channel:  bcChan.ChannelID,
			Port:     bcChan.PortID,
			Retries:  &retries,
			Timeout:  1 * time.Second,
		},
	}

	memo, err := json.Marshal(metadata)
	require.NoError(err)

	_, err = chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName(), transfer, ibc.TransferOptions{
		Memo: string(memo),
	})
	require.NoError(err)

	err = testutil.WaitForBlocks(ctx, waitBlocks, chainA)
	require.NoError(err)

	err = ibcRelayer.Flush(ctx, eRep, pathAB, abChan.ChannelID)
	require.NoError(err)
	time.Sleep(5 * time.Second)

	err = ibcRelayer.Flush(ctx, eRep, pathBC, bcChan.ChannelID)
	require.NoError(err)
	time.Sleep(5 * time.Second)

	// assert balances for user controlled wallets
	chainABalance, err := chainA.GetBalance(ctx, userA.FormattedAddress(), chainA.Config().Denom)
	require.NoError(err)

	chainBBalance, err := chainB.GetBalance(ctx, userB.FormattedAddress(), firstHopIBCDenom)
	require.NoError(err)

	chainCBalance, err := chainC.GetBalance(ctx, timeoutUser.FormattedAddress(), secondHopIBCDenom)
	require.NoError(err)

	firstHopEscrowBalance, err := chainA.GetBalance(ctx, firstHopEscrowAccount, chainA.Config().Denom)
	require.NoError(err)

	secondHopEscrowBalance, err := chainB.GetBalance(ctx, secondHopEscrowAccount, firstHopIBCDenom)
	require.NoError(err)

	fmt.Printf("UserA balance on chain A after: %s (%s)\n", chainABalance.String(), userA.FormattedAddress())
	fmt.Printf("UserB balance on chain B after: %s (%s)\n", chainBBalance.String(), userB.FormattedAddress())
	fmt.Printf("UserC balance on chain C after: %s (%s)\n", chainCBalance.String(), timeoutUser.FormattedAddress())

	fmt.Printf("First hop escrow balance on chain A after: %s (%s)\n", firstHopEscrowBalance.String(), firstHopEscrowAccount)
	fmt.Printf("Second hop escrow balance on chain B after: %s (%s)\n", secondHopEscrowBalance.String(), secondHopEscrowAccount)

	require.True(chainABalance.Equal(initBal.Sub(transferAmount)))
	require.True(chainBBalance.Equal(transferAmount))
	require.True(chainCBalance.Equal(zeroBal))

	expectedFirstHopEscrowAmount = expectedFirstHopEscrowAmount.Add(transferAmount)

	require.True(firstHopEscrowBalance.Equal(expectedFirstHopEscrowAmount))
	require.True(secondHopEscrowBalance.Equal(zeroBal))
}

func (s *TestPFMSuite) TestTimeoutRefundInvalidReceiver() {
	require := s.Require()
	users := interchaintest.GetAndFundTestUsers(s.T(), ctx, s.T().Name(), initBal, chainA)
	timeoutUsers, err := BuildWallets(ctx, s.T().Name(), chainC)
	require.NoError(err)
	userA := users[0]
	userC := timeoutUsers[0]
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
	require.NoError(err)

	_, err = chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName(), transfer,
		ibc.TransferOptions{Memo: string(memo)})
	require.NoError(err)

	err = testutil.WaitForBlocks(ctx, waitBlocks, chainA)
	require.NoError(err)
	err = ibcRelayer.Flush(ctx, eRep, pathAB, abChan.ChannelID)
	require.NoError(err)

	// assert balances for user controlled wallets
	chainABalance, err := chainA.GetBalance(ctx, userA.FormattedAddress(), chainA.Config().Denom)
	require.NoError(err)

	chainBBalance, err := chainB.GetBalance(ctx, userA.FormattedAddress(), firstHopIBCDenom)
	require.NoError(err)

	chainCBalance, err := chainC.GetBalance(ctx, userC.FormattedAddress(), secondHopIBCDenom)
	require.NoError(err)

	require.True(chainABalance.Equal(initBal.Sub(transferAmount)))
	require.True(chainBBalance.Equal(transferAmount))
	require.True(chainCBalance.Equal(zeroBal))

	firstHopEscrowBalance, err := chainA.GetBalance(ctx, firstHopEscrowAccount, chainA.Config().Denom)
	require.NoError(err)

	secondHopEscrowBalance, err := chainB.GetBalance(ctx, secondHopEscrowAccount, firstHopIBCDenom)
	require.NoError(err)

	expectedFirstHopEscrowAmount = expectedFirstHopEscrowAmount.Add(transferAmount)

	require.True(firstHopEscrowBalance.Equal(expectedFirstHopEscrowAmount))
	require.True(secondHopEscrowBalance.Equal(zeroBal))
}

func (s *TestPFMSuite) createVoucherWallets() []ibc.Wallet {
	require := s.Require()
	users := interchaintest.GetAndFundTestUsers(s.T(), ctx, s.T().Name(), initBal, chainA, chainC)
	userA := users[0]
	userC := users[1]
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
	require.NoError(err)

	_, err = chainC.SendIBCTransfer(ctx, cbChan.ChannelID, userC.KeyName(), transfer,
		ibc.TransferOptions{Memo: string(memo)})
	require.NoError(err)

	require.NoError(testutil.WaitForBlocks(ctx, waitBlocks, chainC))
	err = ibcRelayer.Flush(ctx, eRep, pathBC, cbChan.ChannelID)
	require.NoError(err)
	/*
		require.NoError(testutil.WaitForBlocks(ctx, waitBlocks, chainB))
		err = ibcRelayer.Flush(ctx, eRep, pathAB, baChan.ChannelID)
		require.NoError(err)
		require.NoError(testutil.WaitForBlocks(ctx, waitBlocks, chainA))
	*/
	chainCBalance, err := chainC.GetBalance(ctx, userC.FormattedAddress(), configC.Denom)
	require.NoError(err)
	// TODO add balance check for intermediate account on B
	chainABalance, err := chainA.GetBalance(ctx, userA.FormattedAddress(), revSecondHopIBCDenom)
	require.NoError(err)

	chainCExpected := initBal.Sub(transferAmount)
	require.Truef(chainCBalance.Equal(chainCExpected), "chain c expected balance %s, got %s", chainCExpected, chainCBalance)
	require.Truef(chainABalance.Equal(transferAmount), "chain a expected balance %s, got %s", transferAmount,
		chainABalance)

	expectedRevFirstHopEscrowAmount = expectedRevFirstHopEscrowAmount.Add(transferAmount)
	expectedRevSecondHopEscrowAmount = expectedRevSecondHopEscrowAmount.Add(transferAmount)
	s.validateRevEscrow()

	return users
}

func (s *TestPFMSuite) validateRevEscrow() {
	s.T().Helper()
	require := s.Require()

	// assert balances for IBC escrow accounts
	revFirstHopEscrowBalance, err := chainC.GetBalance(ctx, revFirstHopEscrowAccount, configC.Denom)
	require.NoError(err)

	revSecondHopEscrowBalance, err := chainB.GetBalance(ctx, revSecondHopEscrowAccount, revFirstHopIBCDenom)
	require.NoError(err)

	require.Truef(revFirstHopEscrowBalance.Equal(expectedRevFirstHopEscrowAmount), "expected %s got %s",
		expectedRevFirstHopEscrowAmount, revFirstHopEscrowBalance)
	require.Truef(revSecondHopEscrowBalance.Equal(expectedRevSecondHopEscrowAmount), "expected %s got %s",
		expectedRevSecondHopEscrowAmount, revSecondHopEscrowBalance)
}

func (s *TestPFMSuite) TestRevForwardAckErrorInvalidReceiver() {
	require := s.Require()
	users := s.createVoucherWallets()
	userA := users[0]

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
	require.NoError(err)

	_, err = chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName(), transfer,
		ibc.TransferOptions{Memo: string(memo)})
	require.NoError(err)

	err = testutil.WaitForBlocks(ctx, waitBlocks, chainA)
	require.NoError(err)
	err = ibcRelayer.Flush(ctx, eRep, pathAB, abChan.ChannelID)
	require.NoError(err)

	// assert balances for user controlled wallets
	chainABalance, err := chainA.GetBalance(ctx, userA.FormattedAddress(), revSecondHopIBCDenom)
	require.NoError(err)

	// funds should end up in the A user's bech32 transformed address on chain B.
	chainBBalance, err := chainB.GetBalance(ctx, userA.FormattedAddress(), revFirstHopIBCDenom)
	require.NoError(err)

	require.True(chainABalance.Equal(zeroBal))

	// funds should end up in the A user's bech32 transformed address on chain B.
	require.True(chainBBalance.Equal(transferAmount))

	// assert balances for IBC escrow accounts
	revFirstHopEscrowBalance, err := chainC.GetBalance(ctx, revFirstHopEscrowAccount, configC.Denom)
	require.NoError(err)
	require.NoError(err)

	revSecondHopEscrowBalance, err := chainB.GetBalance(ctx, revSecondHopEscrowAccount, revFirstHopIBCDenom)
	require.NoError(err)

	expectedRevSecondHopEscrowAmount = expectedRevSecondHopEscrowAmount.Sub(transferAmount)

	require.True(revFirstHopEscrowBalance.Equal(expectedRevFirstHopEscrowAmount), "expected %s, got %s",
		expectedRevFirstHopEscrowAmount, revFirstHopEscrowBalance)
	require.True(revSecondHopEscrowBalance.Equal(expectedRevSecondHopEscrowAmount),
		"expected %s, got %s", expectedRevSecondHopEscrowAmount, revSecondHopEscrowBalance)
}

func (s *TestPFMSuite) TestRevForwardAckErrorValidReceiver() {
	require := s.Require()
	voucherUsers := s.createVoucherWallets()
	userA := voucherUsers[0]
	users := interchaintest.GetAndFundTestUsers(s.T(), ctx, s.T().Name(), initBal, chainB)
	userB := users[0]
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
	require.NoError(err)

	_, err = chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName(), transfer,
		ibc.TransferOptions{Memo: string(memo)})
	require.NoError(err)

	err = testutil.WaitForBlocks(ctx, waitBlocks, chainA)
	require.NoError(err)

	// assert balances for user controlled wallets
	chainABalance, err := chainA.GetBalance(ctx, userA.FormattedAddress(), revSecondHopIBCDenom)
	require.NoError(err)

	chainBBalance, err := chainB.GetBalance(ctx, userB.FormattedAddress(), revFirstHopIBCDenom)
	require.NoError(err)

	require.True(chainABalance.Equal(zeroBal))
	require.True(chainBBalance.Equal(transferAmount))

	// assert balances for IBC escrow accounts
	revFirstHopEscrowBalance, err := chainC.GetBalance(ctx, revFirstHopEscrowAccount, configC.Denom)
	require.NoError(err)

	revSecondHopEscrowBalance, err := chainB.GetBalance(ctx, revSecondHopEscrowAccount, revFirstHopIBCDenom)
	require.NoError(err)

	expectedRevSecondHopEscrowAmount = expectedRevSecondHopEscrowAmount.Sub(transferAmount)

	require.Truef(revFirstHopEscrowBalance.Equal(expectedRevFirstHopEscrowAmount),
		"expected %s, got %s", expectedRevFirstHopEscrowAmount, revFirstHopEscrowBalance)
	require.Truef(revSecondHopEscrowBalance.Equal(expectedRevSecondHopEscrowAmount),
		"expected %s, got %s", expectedRevSecondHopEscrowAmount, revSecondHopEscrowBalance)
}

func (s *TestPFMSuite) TestRevForwardTimeoutRefundValidReceiver() {
	require := s.Require()
	voucherUsers := s.createVoucherWallets()
	userA := voucherUsers[0]
	users := interchaintest.GetAndFundTestUsers(s.T(), ctx, s.T().Name(), initBal, chainB)
	userB := users[0]
	timeoutUsers, err := BuildWallets(ctx, s.T().Name(), chainC)
	require.NoError(err)
	userC := timeoutUsers[0]
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
	require.NoError(err)

	chainAHeight, err := chainA.Height(ctx)
	require.NoError(err)

	transferTx, err := chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName(), transfer, ibc.TransferOptions{Memo: string(memo)})
	require.NoError(err)

	_, err = testutil.PollForAck(ctx, chainA, chainAHeight, chainAHeight+25, transferTx.Packet)

	require.NoError(err)
	err = testutil.WaitForBlocks(ctx, waitBlocks, chainA)
	require.NoError(err)

	// assert balances for user controlled wallets
	chainABalance, err := chainA.GetBalance(ctx, userA.FormattedAddress(), revSecondHopIBCDenom)
	require.NoError(err)

	chainBBalance, err := chainB.GetBalance(ctx, userB.FormattedAddress(), revFirstHopIBCDenom)
	require.NoError(err)

	chainCBalance, err := chainC.GetBalance(ctx, userC.FormattedAddress(), configC.Denom)
	require.NoError(err)

	require.Truef(chainABalance.Equal(zeroBal), "chain a balance, expected %s, got %s", zeroBal, chainABalance)
	require.Truef(chainBBalance.Equal(transferAmount), "chain b balance, expected %s, got %s", transferAmount, chainBBalance)
	require.Truef(chainCBalance.Equal(zeroBal), "chain c balance, expected %s, got %s", zeroBal, chainCBalance)

	revFirstHopEscrowBalance, err := chainC.GetBalance(ctx, revFirstHopEscrowAccount, configC.Denom)
	require.NoError(err)

	revSecondHopEscrowBalance, err := chainB.GetBalance(ctx, revSecondHopEscrowAccount, revFirstHopIBCDenom)
	require.NoError(err)

	expectedRevSecondHopEscrowAmount = expectedRevSecondHopEscrowAmount.Sub(transferAmount)

	require.Truef(revFirstHopEscrowBalance.Equal(expectedRevFirstHopEscrowAmount),
		"expected %s, got %s", expectedRevFirstHopEscrowAmount, revFirstHopEscrowBalance)
	require.Truef(revSecondHopEscrowBalance.Equal(expectedRevSecondHopEscrowAmount),
		"expected %s, got %s", expectedRevSecondHopEscrowAmount, revSecondHopEscrowBalance)
}

func (s *TestPFMSuite) TestRevForwardTimeoutRefundInvalidReceiver() {
	require := s.Require()
	voucherUsers := s.createVoucherWallets()
	userA := voucherUsers[0]
	users, err := BuildWallets(ctx, s.T().Name(), chainC)
	require.NoError(err)
	userC := users[0]
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
	require.NoError(err)

	_, err = chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName(), transfer, ibc.TransferOptions{
		Memo: string(memo),
	})
	require.NoError(err)
	err = testutil.WaitForBlocks(ctx, waitBlocks, chainA)
	require.NoError(err)
	err = ibcRelayer.Flush(ctx, eRep, pathAB, abChan.ChannelID)
	require.NoError(err)

	// assert balances for user controlled wallets
	chainABalance, err := chainA.GetBalance(ctx, userA.FormattedAddress(), revSecondHopIBCDenom)
	require.NoError(err)

	chainBBalance, err := chainB.GetBalance(ctx, userA.FormattedAddress(), revFirstHopIBCDenom)
	require.NoError(err)

	chainCBalance, err := chainC.GetBalance(ctx, userC.FormattedAddress(), configC.Denom)
	require.NoError(err)

	require.Truef(chainABalance.Equal(zeroBal), "chain a balance, expected %s, got %s", zeroBal, chainABalance)
	require.Truef(chainBBalance.Equal(transferAmount), "chain b balance, expected %s, got %s", transferAmount, chainBBalance)
	require.Truef(chainCBalance.Equal(zeroBal), "chain c balance, expected %s, got %s", zeroBal, chainCBalance)

	revFirstHopEscrowBalance, err := chainC.GetBalance(ctx, revFirstHopEscrowAccount, configC.Denom)
	require.NoError(err)

	revSecondHopEscrowBalance, err := chainB.GetBalance(ctx, revSecondHopEscrowAccount, revFirstHopIBCDenom)
	require.NoError(err)

	expectedRevSecondHopEscrowAmount = expectedRevSecondHopEscrowAmount.Sub(transferAmount)

	require.Truef(revFirstHopEscrowBalance.Equal(expectedRevFirstHopEscrowAmount), "expected %s, got %s", expectedRevFirstHopEscrowAmount, revFirstHopEscrowBalance)
	require.Truef(revSecondHopEscrowBalance.Equal(expectedRevSecondHopEscrowAmount), "expected %s, got %s", expectedRevSecondHopEscrowAmount, revSecondHopEscrowBalance)
}
