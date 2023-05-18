package types_test

import "github.com/cosmos/ibc-apps/modules/async-icq/v7/types"

func (suite *TypesTestSuite) TestContainsQueryPath() {
	allowQueries := []string{
		"path/to/query1",
		"path/to/query2",
	}

	found := types.ContainsQueryPath(allowQueries, "path/to/query1")
	suite.Require().True(found)

	found = types.ContainsQueryPath(allowQueries, "path/to/query2")
	suite.Require().True(found)

	found = types.ContainsQueryPath(allowQueries, "path/to/query3")
	suite.Require().False(found)
}
