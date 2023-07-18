package keeper_test

import (
	"testing"

	"github.com/cosmos/ibc-apps/modules/async-icq/v7/testing/simapp"
	"github.com/cosmos/ibc-apps/modules/async-icq/v7/types"
	"github.com/stretchr/testify/suite"

	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
)

var TestPort = "icq-test"

type KeeperTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
	chainC *ibctesting.TestChain
}

func (suite *KeeperTestSuite) SetupTest() {
	ibctesting.DefaultTestingAppInit = simapp.SetupTestingApp

	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 3)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(1))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(2))
	suite.chainC = suite.coordinator.GetChain(ibctesting.GetChainID(3))
}

func NewICQPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = types.PortID
	path.EndpointB.ChannelConfig.PortID = types.PortID
	path.EndpointA.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointA.ChannelConfig.Version = types.Version
	path.EndpointB.ChannelConfig.Version = types.Version

	return path
}

// SetupICAPath invokes the ICQ entrypoint and subsequent channel handshake handlers
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

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestIsBound() {
	suite.SetupTest()

	path := NewICQPath(suite.chainA, suite.chainB)
	suite.coordinator.SetupConnections(path)

	err := SetupICQPath(path)
	suite.Require().NoError(err)

	isBound := simapp.GetSimApp(suite.chainB).ICQKeeper.IsBound(suite.chainB.GetContext(), path.EndpointB.ChannelConfig.PortID)
	suite.Require().True(isBound)
}
