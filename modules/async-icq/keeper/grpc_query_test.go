package keeper_test

import (
	"github.com/cosmos/ibc-apps/modules/async-icq/v4/testing/simapp"
	"github.com/cosmos/ibc-apps/modules/async-icq/v4/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestQueryParams() {
	ctx := sdk.WrapSDKContext(suite.chainA.GetContext())
	expParams := types.DefaultParams()
	res, _ := simapp.GetSimApp(suite.chainA).ICQKeeper.Params(ctx, &types.QueryParamsRequest{})
	suite.Require().Equal(&expParams, res.Params)
}
