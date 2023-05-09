package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/strangelove-ventures/async-icq/v5/testing/simapp"
	"github.com/strangelove-ventures/async-icq/v5/types"
)

func (suite *KeeperTestSuite) TestQueryParams() {
	ctx := sdk.WrapSDKContext(suite.chainA.GetContext())
	expParams := types.DefaultParams()
	res, _ := simapp.GetSimApp(suite.chainA).ICQKeeper.Params(ctx, &types.QueryParamsRequest{})
	suite.Require().Equal(&expParams, res.Params)
}
