package keeper_test

import (
	"github.com/cosmos/ibc-apps/modules/async-icq/v7/testing/simapp"
	"github.com/cosmos/ibc-apps/modules/async-icq/v7/types"
)

func (suite *KeeperTestSuite) TestInitGenesis() {
	suite.SetupTest()

	genesisState := types.GenesisState{
		HostPort: TestPort,
		Params: types.Params{
			HostEnabled: false,
			AllowQueries: []string{
				"path/to/query1",
				"path/to/query2",
			},
		},
	}

	simapp.GetSimApp(suite.chainA).ICQKeeper.InitGenesis(suite.chainA.GetContext(), genesisState)

	port := simapp.GetSimApp(suite.chainA).ICQKeeper.GetPort(suite.chainA.GetContext())
	suite.Require().Equal(TestPort, port)

	expParams := types.NewParams(
		false,
		[]string{
			"path/to/query1",
			"path/to/query2",
		},
	)
	params := simapp.GetSimApp(suite.chainA).ICQKeeper.GetParams(suite.chainA.GetContext())
	suite.Require().Equal(expParams, params)
}

func (suite *KeeperTestSuite) TestExportGenesis() {
	suite.SetupTest()

	genesisState := simapp.GetSimApp(suite.chainA).ICQKeeper.ExportGenesis(suite.chainA.GetContext())

	suite.Require().Equal(types.PortID, genesisState.GetHostPort())

	expParams := types.DefaultParams()
	suite.Require().Equal(expParams, genesisState.GetParams())
}
