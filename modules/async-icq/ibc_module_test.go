package icq_test

import (
	"testing"

	"github.com/cosmos/gogoproto/proto"
	icq "github.com/cosmos/ibc-apps/modules/async-icq/v8"
	"github.com/cosmos/ibc-apps/modules/async-icq/v8/testing/simapp"
	"github.com/cosmos/ibc-apps/modules/async-icq/v8/types"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	tmprotostate "github.com/cometbft/cometbft/proto/tendermint/state"
	tmstate "github.com/cometbft/cometbft/state"

	clienttypes "github.com/cosmos/ibc-go/v11/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v11/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v11/testing"
)

var (
	// TestVersion defines a reusable icq version string for testing purposes
	TestVersion = "icq-1"

	TestQueryPath = "/store/params/key"
	TestQueryData = "icqhost/HostEnabled"
	version       = "version"
)

type InterchainQueriesTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
}

func TestInterchainQuerySuite(t *testing.T) {
	suite.Run(t, new(InterchainQueriesTestSuite))
}

func (suite *InterchainQueriesTestSuite) SetupTest() {
	ibctesting.DefaultTestingAppInit = simapp.SetupTestingApp

	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(1))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(2))
}

func NewICQPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = types.PortID
	path.EndpointB.ChannelConfig.PortID = types.PortID
	path.EndpointA.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointA.ChannelConfig.Version = TestVersion
	path.EndpointB.ChannelConfig.Version = TestVersion

	return path
}

// SetupICQPath invokes the ICQ entrypoint and subsequent channel handshake handlers
func SetupICQPath(path *ibctesting.Path) error {
	if err := path.EndpointA.ChanOpenInit(); err != nil {
		return err
	}

	if err := path.EndpointB.ChanOpenTry(); err != nil {
		return err
	}

	if err := path.EndpointA.ChanOpenAck(); err != nil {
		return err
	}

	return path.EndpointB.ChanOpenConfirm()
}

func (suite *InterchainQueriesTestSuite) TestOnChanOpenInit() {
	var (
		channel      *channeltypes.Channel
		path         *ibctesting.Path
		counterparty channeltypes.Counterparty
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"success", func() {}, true,
		},
		{
			"empty version string", func() {
				channel.Version = ""
			}, true,
		},
		{
			"invalid order - ORDERED", func() {
				channel.Ordering = channeltypes.ORDERED
			}, false,
		},
		{
			"invalid port ID", func() {
				path.EndpointA.ChannelConfig.PortID = ibctesting.MockPort
			}, false,
		},
		{
			"invalid version", func() {
				channel.Version = version
			}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			path = NewICQPath(suite.chainA, suite.chainB)
			suite.coordinator.SetupConnections(path)
			path.EndpointA.ChannelID = ibctesting.FirstChannelID

			counterparty = channeltypes.NewCounterparty(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)
			channel = &channeltypes.Channel{
				State:          channeltypes.INIT,
				Ordering:       channeltypes.UNORDERED,
				Counterparty:   counterparty,
				ConnectionHops: []string{path.EndpointA.ConnectionID},
				Version:        types.Version,
			}

			tc.malleate() // explicitly change fields in channel and testChannel

			icqModule := icq.NewIBCModule(simapp.GetSimApp(suite.chainA).ICQKeeper)
			version, err := icqModule.OnChanOpenInit(suite.chainA.GetContext(), channel.Ordering, channel.ConnectionHops,
				path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, counterparty, channel.Version,
			)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(types.Version, version)
			} else {
				suite.Require().Error(err)
				suite.Require().Equal(version, "")
			}
		})
	}
}

func (suite *InterchainQueriesTestSuite) TestOnChanOpenTry() {
	var (
		channel             *channeltypes.Channel
		path                *ibctesting.Path
		counterparty        channeltypes.Counterparty
		counterpartyVersion string
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"success", func() {}, true,
		},
		{
			"invalid order - ORDERED", func() {
				channel.Ordering = channeltypes.ORDERED
			}, false,
		},
		{
			"invalid port ID", func() {
				path.EndpointA.ChannelConfig.PortID = ibctesting.MockPort
			}, false,
		},
		{
			"invalid counterparty version", func() {
				counterpartyVersion = version
			}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			path = NewICQPath(suite.chainA, suite.chainB)
			suite.coordinator.SetupConnections(path)
			path.EndpointA.ChannelID = ibctesting.FirstChannelID

			counterparty = channeltypes.NewCounterparty(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)
			channel = &channeltypes.Channel{
				State:          channeltypes.TRYOPEN,
				Ordering:       channeltypes.UNORDERED,
				Counterparty:   counterparty,
				ConnectionHops: []string{path.EndpointA.ConnectionID},
				Version:        types.Version,
			}
			counterpartyVersion = types.Version

			tc.malleate() // explicitly change fields in channel and testChannel

			icqModule := icq.NewIBCModule(simapp.GetSimApp(suite.chainA).ICQKeeper)
			version, err := icqModule.OnChanOpenTry(suite.chainA.GetContext(), channel.Ordering, channel.ConnectionHops,
				path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, channel.Counterparty, counterpartyVersion,
			)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(types.Version, version)
			} else {
				suite.Require().Error(err)
				suite.Require().Equal("", version)
			}
		})
	}
}

func (suite *InterchainQueriesTestSuite) TestOnChanOpenAck() {
	var counterpartyVersion string

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"success", func() {}, true,
		},
		{
			"invalid counterparty version", func() {
				counterpartyVersion = version
			}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			path := NewICQPath(suite.chainA, suite.chainB)
			suite.coordinator.SetupConnections(path)
			path.EndpointA.ChannelID = ibctesting.FirstChannelID
			counterpartyVersion = types.Version

			tc.malleate() // explicitly change fields in channel and testChannel

			icqModule := icq.NewIBCModule(simapp.GetSimApp(suite.chainA).ICQKeeper)
			err := icqModule.OnChanOpenAck(suite.chainA.GetContext(), path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointA.Counterparty.ChannelID, counterpartyVersion)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *InterchainQueriesTestSuite) TestOnAcknowledgementPacket() {
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"icq OnAcknowledgementPacket fails with ErrInvalidChannelFlow", func() {}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			path := NewICQPath(suite.chainA, suite.chainB)
			suite.coordinator.SetupConnections(path)

			err := SetupICQPath(path)
			suite.Require().NoError(err)

			tc.malleate() // malleate mutates test data

			packet := channeltypes.NewPacket(
				[]byte("empty packet data"),
				suite.chainA.SenderAccount.GetSequence(),
				path.EndpointB.ChannelConfig.PortID,
				path.EndpointB.ChannelID,
				path.EndpointA.ChannelConfig.PortID,
				path.EndpointA.ChannelID,
				clienttypes.NewHeight(0, 100),
				0,
			)

			icqModule := icq.NewIBCModule(simapp.GetSimApp(suite.chainB).ICQKeeper)
			err = icqModule.OnAcknowledgementPacket(suite.chainB.GetContext(), types.Version, packet, []byte("ackBytes"), nil)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *InterchainQueriesTestSuite) TestOnTimeoutPacket() {
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"icq OnTimeoutPacket fails with ErrInvalidChannelFlow", func() {}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			path := NewICQPath(suite.chainA, suite.chainB)
			suite.coordinator.SetupConnections(path)

			err := SetupICQPath(path)
			suite.Require().NoError(err)

			tc.malleate() // malleate mutates test data

			packet := channeltypes.NewPacket(
				[]byte("empty packet data"),
				suite.chainA.SenderAccount.GetSequence(),
				path.EndpointB.ChannelConfig.PortID,
				path.EndpointB.ChannelID,
				path.EndpointA.ChannelConfig.PortID,
				path.EndpointA.ChannelID,
				clienttypes.NewHeight(0, 100),
				0,
			)

			icqModule := icq.NewIBCModule(simapp.GetSimApp(suite.chainA).ICQKeeper)
			err = icqModule.OnTimeoutPacket(suite.chainA.GetContext(), types.Version, packet, nil)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// The safety of including SDK MsgResponses in the acknowledgement rests
// on the inclusion of the abcitypes.ResponseDeliverTx.Data in the
// abcitypes.ResposneDeliverTx hash. If the abcitypes.ResponseDeliverTx.Data
// gets removed from consensus they must no longer be used in the packet
// acknowledgement.
//
// This test acts as an indicqtor that the abcitypes.ResponseDeliverTx.Data
// may no longer be deterministic.
func (suite *InterchainQueriesTestSuite) TestABCICodeDeterminism() {
	msgResponseBz, err := proto.Marshal(&channeltypes.MsgChannelOpenInitResponse{})
	suite.Require().NoError(err)

	msgData := &sdk.MsgData{ //nolint:staticcheck
		MsgType: sdk.MsgTypeURL(&channeltypes.MsgChannelOpenInit{}),
		Data:    msgResponseBz,
	}

	txResponse, err := proto.Marshal(&sdk.TxMsgData{
		Data: []*sdk.MsgData{msgData}, //nolint:staticcheck
	})
	suite.Require().NoError(err)

	deliverTx := abcitypes.ExecTxResult{
		Data: txResponse,
	}
	responses := tmprotostate.LegacyABCIResponses{
		DeliverTxs: []*abcitypes.ExecTxResult{
			&deliverTx,
		},
	}

	differentMsgResponseBz, err := proto.Marshal(&channeltypes.MsgRecvPacketResponse{})
	suite.Require().NoError(err)

	differentMsgData := &sdk.MsgData{ //nolint:staticcheck
		MsgType: sdk.MsgTypeURL(&channeltypes.MsgRecvPacket{}),
		Data:    differentMsgResponseBz,
	}

	differentTxResponse, err := proto.Marshal(&sdk.TxMsgData{
		Data: []*sdk.MsgData{differentMsgData}, //nolint:staticcheck
	})
	suite.Require().NoError(err)

	differentDeliverTx := abcitypes.ExecTxResult{
		Data: differentTxResponse,
	}

	differentResponses := tmprotostate.LegacyABCIResponses{
		DeliverTxs: []*abcitypes.ExecTxResult{
			&differentDeliverTx,
		},
	}

	hash := tmstate.TxResultsHash(responses.DeliverTxs)
	differentHash := tmstate.TxResultsHash(differentResponses.DeliverTxs)

	suite.Require().NotEqual(hash, differentHash)
}
