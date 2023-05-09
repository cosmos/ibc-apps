package keeper_test

import (
	"github.com/strangelove-ventures/async-icq/v5/testing/simapp"
	"github.com/strangelove-ventures/async-icq/v5/types"
)

func (suite *KeeperTestSuite) TestParams() {
	expParams := types.DefaultParams()

	params := simapp.GetSimApp(suite.chainA).ICQKeeper.GetParams(suite.chainA.GetContext())
	suite.Require().Equal(expParams, params)

	expParams.HostEnabled = false
	expParams.AllowQueries = []string{"/cosmos.staking.v1beta1.MsgDelegate"}
	simapp.GetSimApp(suite.chainA).ICQKeeper.SetParams(suite.chainA.GetContext(), expParams)
	params = simapp.GetSimApp(suite.chainA).ICQKeeper.GetParams(suite.chainA.GetContext())
	suite.Require().Equal(expParams, params)
}
