package keeper_test

import (
	"fmt"

	"github.com/cosmos/ibc-apps/modules/async-icq/v7/testing/simapp"
	"github.com/cosmos/ibc-apps/modules/async-icq/v7/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	abcitypes "github.com/cometbft/cometbft/abci/types"

	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
)

func (suite *KeeperTestSuite) TestOnRecvPacket() {
	var (
		path       *ibctesting.Path
		packetData []byte
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"icq successfully queries banktypes.AllBalances",
			func() {
				q := banktypes.QueryAllBalancesRequest{
					Address: suite.chainA.SenderAccount.GetAddress().String(),
					Pagination: &query.PageRequest{
						Offset: 0,
						Limit:  10,
					},
				}
				reqs := []abcitypes.RequestQuery{
					{
						Path: "/cosmos.bank.v1beta1.Query/AllBalances",
						Data: simapp.GetSimApp(suite.chainA).AppCodec().MustMarshal(&q),
					},
				}
				data, err := types.SerializeCosmosQuery(reqs)
				suite.Require().NoError(err)

				icqPacketData := types.InterchainQueryPacketData{
					Data: data,
				}
				packetData = icqPacketData.GetBytes()

				params := types.NewParams(true, []string{"/cosmos.bank.v1beta1.Query/AllBalances"})
				if err := simapp.GetSimApp(suite.chainB).ICQKeeper.SetParams(suite.chainB.GetContext(), params); err != nil {
					panic(err)
				}
			},
			true,
		},
		{
			"cannot unmarshal interchain query packet data",
			func() {
				packetData = []byte{}
			},
			false,
		},
		{
			"cannot deserialize interchain query packet data messages",
			func() {
				data := []byte("invalid packet data")

				icaPacketData := types.InterchainQueryPacketData{
					Data: data,
				}

				packetData = icaPacketData.GetBytes()
			},
			false,
		},
		{
			"unauthorised: message type not allowed", // NOTE: do not update params to explicitly force the error
			func() {
				q := banktypes.QueryAllBalancesRequest{}
				reqs := []abcitypes.RequestQuery{
					{
						Path: "/cosmos.bank.v1beta1.Query/AllBalances",
						Data: simapp.GetSimApp(suite.chainA).AppCodec().MustMarshal(&q),
					},
				}
				data, err := types.SerializeCosmosQuery(reqs)
				suite.Require().NoError(err)

				icaPacketData := types.InterchainQueryPacketData{
					Data: data,
				}
				packetData = icaPacketData.GetBytes()
			},
			false,
		},
		{
			"unauthorised: can not perform historical query (i.e. height != 0)",
			func() {
				q := banktypes.QueryAllBalancesRequest{}
				reqs := []abcitypes.RequestQuery{
					{
						Path:   "/cosmos.bank.v1beta1.Query/AllBalances",
						Data:   simapp.GetSimApp(suite.chainA).AppCodec().MustMarshal(&q),
						Height: 1,
					},
				}
				data, err := types.SerializeCosmosQuery(reqs)
				suite.Require().NoError(err)

				icaPacketData := types.InterchainQueryPacketData{
					Data: data,
				}
				packetData = icaPacketData.GetBytes()

				params := types.NewParams(true, []string{"/cosmos.bank.v1beta1.Query/AllBalances"})
				if err := simapp.GetSimApp(suite.chainB).ICQKeeper.SetParams(suite.chainB.GetContext(), params); err != nil {
					panic(err)
				}
			},
			false,
		},
		{
			"unauthorised: can not fetch query proof (i.e. prove == true)",
			func() {
				q := banktypes.QueryAllBalancesRequest{}
				reqs := []abcitypes.RequestQuery{
					{
						Path:  "/cosmos.bank.v1beta1.Query/AllBalances",
						Data:  simapp.GetSimApp(suite.chainA).AppCodec().MustMarshal(&q),
						Prove: true,
					},
				}
				data, err := types.SerializeCosmosQuery(reqs)
				suite.Require().NoError(err)

				icaPacketData := types.InterchainQueryPacketData{
					Data: data,
				}
				packetData = icaPacketData.GetBytes()

				params := types.NewParams(true, []string{"/cosmos.bank.v1beta1.Query/AllBalances"})
				if err := simapp.GetSimApp(suite.chainB).ICQKeeper.SetParams(suite.chainB.GetContext(), params); err != nil {
					panic(err)
				}
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.msg, func() {
			suite.SetupTest() // reset

			path = NewICQPath(suite.chainA, suite.chainB)
			suite.coordinator.SetupConnections(path)

			err := SetupICQPath(path)
			suite.Require().NoError(err)

			tc.malleate() // malleate mutates test data

			packet := channeltypes.NewPacket(
				packetData,
				suite.chainA.SenderAccount.GetSequence(),
				path.EndpointA.ChannelConfig.PortID,
				path.EndpointA.ChannelID,
				path.EndpointB.ChannelConfig.PortID,
				path.EndpointB.ChannelID,
				clienttypes.NewHeight(1, 100),
				0,
			)

			txResponse, err := simapp.GetSimApp(suite.chainB).ICQKeeper.OnRecvPacket(suite.chainB.GetContext(), packet)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(txResponse)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(txResponse)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestOutOfGasOnSlowQueries() {
	path := NewICQPath(suite.chainA, suite.chainB)
	suite.coordinator.SetupConnections(path)

	err := SetupICQPath(path)
	suite.Require().NoError(err)

	q := banktypes.QueryAllBalancesRequest{
		Address: suite.chainB.SenderAccount.GetAddress().String(),
		Pagination: &query.PageRequest{
			Offset: 0,
			Limit:  100_000_000,
		},
	}
	reqs := []abcitypes.RequestQuery{
		{
			Path: "/cosmos.bank.v1beta1.Query/AllBalances",
			Data: simapp.GetSimApp(suite.chainB).AppCodec().MustMarshal(&q),
		},
	}
	data, err := types.SerializeCosmosQuery(reqs)
	suite.Require().NoError(err)

	icqPacketData := types.InterchainQueryPacketData{
		Data: data,
	}
	packetData := icqPacketData.GetBytes()

	params := types.NewParams(true, []string{"/cosmos.bank.v1beta1.Query/AllBalances"})
	if err := simapp.GetSimApp(suite.chainB).ICQKeeper.SetParams(suite.chainB.GetContext(), params); err != nil {
		panic(err)
	}

	packet := channeltypes.NewPacket(
		packetData,
		suite.chainA.SenderAccount.GetSequence(),
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		clienttypes.NewHeight(1, 100),
		0,
	)

	ctx := suite.chainB.GetContext()
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(2000))
	// enough gas for this small query, but not for the larger one. This one should work
	_, err = simapp.GetSimApp(suite.chainB).ICQKeeper.OnRecvPacket(ctx, packet)
	suite.Require().NoError(err)

	// fund account with 10_000 denoms
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	for i := 0; i < 10_000; i++ {
		denom := fmt.Sprintf("denom%d", i)
		err = simapp.GetSimApp(suite.chainB).BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(denom, 10)))
		suite.Require().NoError(err)
		err = simapp.GetSimApp(suite.chainB).BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, suite.chainB.SenderAccount.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin(denom, 10)))
		suite.Require().NoError(err)

	}

	// We need to call NextBlock() so that the context is committed. This doesn't matter as much anymore,
	// but previous versions didn't pass the context to the query, so the test would've picked the previous
	// block's data
	suite.chainB.NextBlock()
	ctx = suite.chainB.GetContext()

	packet = channeltypes.NewPacket(
		packetData,
		suite.chainA.SenderAccount.GetSequence(),
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		clienttypes.NewHeight(1, 100),
		0,
	)
	//

	// and this one should panic
	suite.Assert().Panics(func() {
		ctx = ctx.WithGasMeter(sdk.NewGasMeter(2000))
		_, _ = simapp.GetSimApp(suite.chainB).ICQKeeper.OnRecvPacket(ctx, packet)
	}, "out of gas")
}
