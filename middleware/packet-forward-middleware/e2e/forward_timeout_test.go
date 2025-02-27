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

func TestTimeoutOnForward(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	var (
		ctx                                    = context.Background()
		client, network                        = interchaintest.DockerSetup(t)
		rep                                    = testreporter.NewNopReporter()
		eRep                                   = rep.RelayerExecReporter(t)
		chainIdA, chainIdB, chainIdC, chainIdD = "chain-a", "chain-b", "chain-c", "chain-d"
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

	baseCfg.ChainID = chainIdD
	configD := baseCfg

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{Name: "pfm", ChainConfig: configA, NumFullNodes: &fullNodes, NumValidators: &vals},
		{Name: "pfm", ChainConfig: configB, NumFullNodes: &fullNodes, NumValidators: &vals},
		{Name: "pfm", ChainConfig: configC, NumFullNodes: &fullNodes, NumValidators: &vals},
		{Name: "pfm", ChainConfig: configD, NumFullNodes: &fullNodes, NumValidators: &vals},
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
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), initBal, chainA, chainB, chainC, chainD)

	abChan, err := ibc.GetTransferChannel(ctx, r, eRep, chainIdA, chainIdB)
	require.NoError(t, err)

	baChan := abChan.Counterparty

	cbChan, err := ibc.GetTransferChannel(ctx, r, eRep, chainIdC, chainIdB)
	require.NoError(t, err)

	bcChan := cbChan.Counterparty

	dcChan, err := ibc.GetTransferChannel(ctx, r, eRep, chainIdD, chainIdC)
	require.NoError(t, err)

	cdChan := dcChan.Counterparty

	userA, userB, userC, userD := users[0], users[1], users[2], users[3]

	// Compose the prefixed denoms and ibc denom for asserting balances
	firstHopDenom := transfertypes.GetPrefixedDenom(baChan.PortID, baChan.ChannelID, chainA.Config().Denom)
	secondHopDenom := transfertypes.GetPrefixedDenom(cbChan.PortID, cbChan.ChannelID, firstHopDenom)
	thirdHopDenom := transfertypes.GetPrefixedDenom(dcChan.PortID, dcChan.ChannelID, secondHopDenom)

	firstHopDenomTrace := transfertypes.ParseDenomTrace(firstHopDenom)
	secondHopDenomTrace := transfertypes.ParseDenomTrace(secondHopDenom)
	thirdHopDenomTrace := transfertypes.ParseDenomTrace(thirdHopDenom)

	firstHopIBCDenom := firstHopDenomTrace.IBCDenom()
	secondHopIBCDenom := secondHopDenomTrace.IBCDenom()
	thirdHopIBCDenom := thirdHopDenomTrace.IBCDenom()

	firstHopEscrowAccount := transfertypes.GetEscrowAddress(abChan.PortID, abChan.ChannelID).String()
	secondHopEscrowAccount := transfertypes.GetEscrowAddress(bcChan.PortID, bcChan.ChannelID).String()
	thirdHopEscrowAccount := transfertypes.GetEscrowAddress(cdChan.PortID, abChan.ChannelID).String()

	zeroBal := math.ZeroInt()
	transferAmount := math.NewInt(100_000)

	// Attempt to send packet from Chain A->Chain B->Chain C->Chain D
	transfer := ibc.WalletAmount{
		Address: userB.FormattedAddress(),
		Denom:   chainA.Config().Denom,
		Amount:  transferAmount,
	}

	retries := uint8(0)
	secondHopMetadata := &PacketMetadata{
		Forward: &ForwardMetadata{
			Receiver: userD.FormattedAddress(),
			Channel:  cdChan.ChannelID,
			Port:     cdChan.PortID,
			Retries:  &retries,
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
			Retries:  &retries,
			Timeout:  time.Second * 10, // Set low timeout for forward from chainB<>chainC
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
	err = r.StartRelayer(ctx, eRep, pathAB, pathBC, pathCD)
	require.NoError(t, err)

	chainAHeight, err := chainA.Height(ctx)
	require.NoError(t, err)

	chainBHeight, err = chainB.Height(ctx)
	require.NoError(t, err)

	// Poll for the MsgTimeout on chainB and the MsgAck on chainA
	_, err = cosmos.PollForMessage[*chantypes.MsgTimeout](ctx, chainB, chainB.Config().EncodingConfig.InterfaceRegistry, chainBHeight, chainBHeight+20, nil)
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

	chainCBalance, err := chainC.GetBalance(ctx, userC.FormattedAddress(), secondHopIBCDenom)
	require.NoError(t, err)

	chainDBalance, err := chainD.GetBalance(ctx, userD.FormattedAddress(), thirdHopIBCDenom)
	require.NoError(t, err)

	require.True(t, chainABalance.Equal(initBal))
	require.True(t, chainBBalance.Equal(zeroBal))
	require.True(t, chainCBalance.Equal(zeroBal))
	require.True(t, chainDBalance.Equal(zeroBal))

	firstHopEscrowBalance, err := chainA.GetBalance(ctx, firstHopEscrowAccount, chainA.Config().Denom)
	require.NoError(t, err)

	secondHopEscrowBalance, err := chainB.GetBalance(ctx, secondHopEscrowAccount, firstHopIBCDenom)
	require.NoError(t, err)

	thirdHopEscrowBalance, err := chainC.GetBalance(ctx, thirdHopEscrowAccount, secondHopIBCDenom)
	require.NoError(t, err)

	require.True(t, firstHopEscrowBalance.Equal(zeroBal))
	require.True(t, secondHopEscrowBalance.Equal(zeroBal))
	require.True(t, thirdHopEscrowBalance.Equal(zeroBal))

	// Send IBC transfer from ChainA -> ChainB -> ChainC -> ChainD that will succeed
	secondHopMetadata = &PacketMetadata{
		Forward: &ForwardMetadata{
			Receiver: userD.FormattedAddress(),
			Channel:  cdChan.ChannelID,
			Port:     cdChan.PortID,
		},
	}
	nextBz, err = json.Marshal(secondHopMetadata)
	require.NoError(t, err)
	next = string(nextBz)

	firstHopMetadata = &PacketMetadata{
		Forward: &ForwardMetadata{
			Receiver: userC.FormattedAddress(),
			Channel:  bcChan.ChannelID,
			Port:     bcChan.PortID,
			Next:     &next,
		},
	}

	memo, err = json.Marshal(firstHopMetadata)
	require.NoError(t, err)

	opts = ibc.TransferOptions{
		Memo: string(memo),
	}

	chainAHeight, err = chainA.Height(ctx)
	require.NoError(t, err)

	transferTx, err = chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName(), transfer, opts)
	require.NoError(t, err)

	_, err = testutil.PollForAck(ctx, chainA, chainAHeight, chainAHeight+30, transferTx.Packet)
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 10, chainA)
	require.NoError(t, err)

	// Assert balances are updated to reflect tokens now being on ChainD
	chainABalance, err = chainA.GetBalance(ctx, userA.FormattedAddress(), chainA.Config().Denom)
	require.NoError(t, err)

	chainBBalance, err = chainB.GetBalance(ctx, userB.FormattedAddress(), firstHopIBCDenom)
	require.NoError(t, err)

	chainCBalance, err = chainC.GetBalance(ctx, userC.FormattedAddress(), secondHopIBCDenom)
	require.NoError(t, err)

	chainDBalance, err = chainD.GetBalance(ctx, userD.FormattedAddress(), thirdHopIBCDenom)
	require.NoError(t, err)

	require.True(t, chainABalance.Equal(initBal.Sub(transferAmount)))
	require.True(t, chainBBalance.Equal(zeroBal))
	require.True(t, chainCBalance.Equal(zeroBal))
	require.True(t, chainDBalance.Equal(transferAmount))

	firstHopEscrowBalance, err = chainA.GetBalance(ctx, firstHopEscrowAccount, chainA.Config().Denom)
	require.NoError(t, err)

	secondHopEscrowBalance, err = chainB.GetBalance(ctx, secondHopEscrowAccount, firstHopIBCDenom)
	require.NoError(t, err)

	thirdHopEscrowBalance, err = chainC.GetBalance(ctx, thirdHopEscrowAccount, secondHopIBCDenom)
	require.NoError(t, err)

	require.True(t, firstHopEscrowBalance.Equal(transferAmount))
	require.True(t, secondHopEscrowBalance.Equal(transferAmount))
	require.True(t, thirdHopEscrowBalance.Equal(transferAmount))

	// Compose IBC tx that will attempt to go from ChainD -> ChainC -> ChainB -> ChainA but timeout between ChainB->ChainA
	transfer = ibc.WalletAmount{
		Address: userC.FormattedAddress(),
		Denom:   thirdHopDenom,
		Amount:  transferAmount,
	}

	secondHopMetadata = &PacketMetadata{
		Forward: &ForwardMetadata{
			Receiver: userA.FormattedAddress(),
			Channel:  baChan.ChannelID,
			Port:     baChan.PortID,
			Timeout:  1 * time.Second,
		},
	}
	nextBz, err = json.Marshal(secondHopMetadata)
	require.NoError(t, err)
	next = string(nextBz)

	firstHopMetadata = &PacketMetadata{
		Forward: &ForwardMetadata{
			Receiver: userB.FormattedAddress(),
			Channel:  cbChan.ChannelID,
			Port:     cbChan.PortID,
			Next:     &next,
		},
	}

	memo, err = json.Marshal(firstHopMetadata)
	require.NoError(t, err)

	chainDHeight, err := chainD.Height(ctx)
	require.NoError(t, err)

	transferTx, err = chainD.SendIBCTransfer(ctx, dcChan.ChannelID, userD.KeyName(), transfer, ibc.TransferOptions{Memo: string(memo)})
	require.NoError(t, err)

	_, err = testutil.PollForAck(ctx, chainD, chainDHeight, chainDHeight+25, transferTx.Packet)
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 5, chainD)
	require.NoError(t, err)

	// Assert balances to ensure timeout happened and user funds are still present on ChainD
	chainABalance, err = chainA.GetBalance(ctx, userA.FormattedAddress(), chainA.Config().Denom)
	require.NoError(t, err)

	chainBBalance, err = chainB.GetBalance(ctx, userB.FormattedAddress(), firstHopIBCDenom)
	require.NoError(t, err)

	chainCBalance, err = chainC.GetBalance(ctx, userC.FormattedAddress(), secondHopIBCDenom)
	require.NoError(t, err)

	chainDBalance, err = chainD.GetBalance(ctx, userD.FormattedAddress(), thirdHopIBCDenom)
	require.NoError(t, err)

	require.True(t, chainABalance.Equal(initBal.Sub(transferAmount)))
	require.True(t, chainBBalance.Equal(zeroBal))
	require.True(t, chainCBalance.Equal(zeroBal))
	require.True(t, chainDBalance.Equal(transferAmount))

	firstHopEscrowBalance, err = chainA.GetBalance(ctx, firstHopEscrowAccount, chainA.Config().Denom)
	require.NoError(t, err)

	secondHopEscrowBalance, err = chainB.GetBalance(ctx, secondHopEscrowAccount, firstHopIBCDenom)
	require.NoError(t, err)

	thirdHopEscrowBalance, err = chainC.GetBalance(ctx, thirdHopEscrowAccount, secondHopIBCDenom)
	require.NoError(t, err)

	require.True(t, firstHopEscrowBalance.Equal(transferAmount))
	require.True(t, secondHopEscrowBalance.Equal(transferAmount))
	require.True(t, thirdHopEscrowBalance.Equal(transferAmount))

	// ---

	// Compose IBC tx that will go from ChainD -> ChainC -> ChainB -> ChainA and succeed.
	transfer = ibc.WalletAmount{
		Address: userC.FormattedAddress(),
		Denom:   thirdHopDenom,
		Amount:  transferAmount,
	}

	secondHopMetadata = &PacketMetadata{
		Forward: &ForwardMetadata{
			Receiver: userA.FormattedAddress(),
			Channel:  baChan.ChannelID,
			Port:     baChan.PortID,
		},
	}
	nextBz, err = json.Marshal(secondHopMetadata)
	require.NoError(t, err)
	next = string(nextBz)

	firstHopMetadata = &PacketMetadata{
		Forward: &ForwardMetadata{
			Receiver: userB.FormattedAddress(),
			Channel:  cbChan.ChannelID,
			Port:     cbChan.PortID,
			Next:     &next,
		},
	}

	memo, err = json.Marshal(firstHopMetadata)
	require.NoError(t, err)

	chainDHeight, err = chainD.Height(ctx)
	require.NoError(t, err)

	transferTx, err = chainD.SendIBCTransfer(ctx, dcChan.ChannelID, userD.KeyName(), transfer, ibc.TransferOptions{Memo: string(memo)})
	require.NoError(t, err)

	_, err = testutil.PollForAck(ctx, chainD, chainDHeight, chainDHeight+25, transferTx.Packet)
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 5, chainD)
	require.NoError(t, err)

	// Assert balances to ensure timeout happened and user funds are still present on ChainD
	chainABalance, err = chainA.GetBalance(ctx, userA.FormattedAddress(), chainA.Config().Denom)
	require.NoError(t, err)

	chainBBalance, err = chainB.GetBalance(ctx, userB.FormattedAddress(), firstHopIBCDenom)
	require.NoError(t, err)

	chainCBalance, err = chainC.GetBalance(ctx, userC.FormattedAddress(), secondHopIBCDenom)
	require.NoError(t, err)

	chainDBalance, err = chainD.GetBalance(ctx, userD.FormattedAddress(), thirdHopIBCDenom)
	require.NoError(t, err)

	require.True(t, chainABalance.Equal(initBal))
	require.True(t, chainBBalance.Equal(zeroBal))
	require.True(t, chainCBalance.Equal(zeroBal))
	require.True(t, chainDBalance.Equal(zeroBal))

	firstHopEscrowBalance, err = chainA.GetBalance(ctx, firstHopEscrowAccount, chainA.Config().Denom)
	require.NoError(t, err)

	secondHopEscrowBalance, err = chainB.GetBalance(ctx, secondHopEscrowAccount, firstHopIBCDenom)
	require.NoError(t, err)

	thirdHopEscrowBalance, err = chainC.GetBalance(ctx, thirdHopEscrowAccount, secondHopIBCDenom)
	require.NoError(t, err)

	require.True(t, firstHopEscrowBalance.Equal(zeroBal))
	require.True(t, secondHopEscrowBalance.Equal(zeroBal))
	require.True(t, thirdHopEscrowBalance.Equal(zeroBal))

	// ----- 2

	// Compose IBC tx that will go from ChainD -> ChainC -> ChainB -> ChainA and succeed.
	transfer = ibc.WalletAmount{
		Address: userB.FormattedAddress(),
		Denom:   chainA.Config().Denom,
		Amount:  transferAmount,
	}

	firstHopMetadata = &PacketMetadata{
		Forward: &ForwardMetadata{
			Receiver: userA.FormattedAddress(),
			Channel:  baChan.ChannelID,
			Port:     baChan.PortID,
			Timeout:  1 * time.Second,
		},
	}

	memo, err = json.Marshal(firstHopMetadata)
	require.NoError(t, err)

	chainAHeight, err = chainA.Height(ctx)
	require.NoError(t, err)

	transferTx, err = chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName(), transfer, ibc.TransferOptions{Memo: string(memo)})
	require.NoError(t, err)

	_, err = testutil.PollForAck(ctx, chainA, chainAHeight, chainAHeight+25, transferTx.Packet)
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 5, chainA)
	require.NoError(t, err)

	// Assert balances to ensure timeout happened and user funds are still present on ChainD
	chainABalance, err = chainA.GetBalance(ctx, userA.FormattedAddress(), chainA.Config().Denom)
	require.NoError(t, err)

	chainBBalance, err = chainB.GetBalance(ctx, userB.FormattedAddress(), firstHopIBCDenom)
	require.NoError(t, err)

	chainCBalance, err = chainC.GetBalance(ctx, userC.FormattedAddress(), secondHopIBCDenom)
	require.NoError(t, err)

	chainDBalance, err = chainD.GetBalance(ctx, userD.FormattedAddress(), thirdHopIBCDenom)
	require.NoError(t, err)

	require.True(t, chainABalance.Equal(initBal))
	require.True(t, chainBBalance.Equal(zeroBal))
	require.True(t, chainCBalance.Equal(zeroBal))
	require.True(t, chainDBalance.Equal(zeroBal))

	firstHopEscrowBalance, err = chainA.GetBalance(ctx, firstHopEscrowAccount, chainA.Config().Denom)
	require.NoError(t, err)

	secondHopEscrowBalance, err = chainB.GetBalance(ctx, secondHopEscrowAccount, firstHopIBCDenom)
	require.NoError(t, err)

	thirdHopEscrowBalance, err = chainC.GetBalance(ctx, thirdHopEscrowAccount, secondHopIBCDenom)
	require.NoError(t, err)

	require.True(t, firstHopEscrowBalance.Equal(zeroBal))
	require.True(t, secondHopEscrowBalance.Equal(zeroBal))
	require.True(t, thirdHopEscrowBalance.Equal(zeroBal))
}
