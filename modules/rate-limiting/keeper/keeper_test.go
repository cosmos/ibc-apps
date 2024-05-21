package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/ibc-apps/modules/rate-limiting/v8/testing/simapp/apptesting"
	"github.com/cosmos/ibc-apps/modules/rate-limiting/v8/types"
)

type KeeperTestSuite struct {
	apptesting.AppTestHelper
	QueryClient types.QueryClient
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	s.QueryClient = types.NewQueryClient(s.QueryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
