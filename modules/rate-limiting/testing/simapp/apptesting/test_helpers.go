package apptesting

import (
	app "github.com/cosmos/ibc-apps/modules/rate-limiting/v10/testing/simapp"
	"github.com/stretchr/testify/suite"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cometbft/cometbft/v2/crypto/ed25519"
)

var TestChainId = "chain-0"

type AppTestHelper struct {
	suite.Suite

	App         *app.SimApp
	QueryHelper *baseapp.QueryServiceTestHelper
	TestAccs    []sdk.AccAddress
	Ctx         sdk.Context
}

// AppTestHelper Constructor
func (s *AppTestHelper) Setup() {
	s.App = app.InitTestingApp()
	s.Ctx = s.App.BaseApp.NewContext(false)
	s.QueryHelper = &baseapp.QueryServiceTestHelper{
		GRPCQueryRouter: s.App.GRPCQueryRouter(),
		Ctx:             s.Ctx,
	}
	s.TestAccs = CreateRandomAccounts(3)
}

// Generate random account addresses
func CreateRandomAccounts(numAccts int) []sdk.AccAddress {
	testAddrs := make([]sdk.AccAddress, numAccts)
	for i := 0; i < numAccts; i++ {
		pk := ed25519.GenPrivKey().PubKey()
		testAddrs[i] = sdk.AccAddress(pk.Address())
	}

	return testAddrs
}

// Helper function to confirm the upgrade block processes without error
func (s *AppTestHelper) ConfirmUpgradeSucceededs(upgradeName string, upgradeHeight int64) {
	s.Ctx = s.Ctx.WithBlockHeight(upgradeHeight - 1)
	plan := upgradetypes.Plan{
		Name:   upgradeName,
		Height: upgradeHeight,
	}

	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, err = s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().NoError(err)

	s.Ctx = s.Ctx.WithBlockHeight(upgradeHeight)
	s.Require().NotPanics(func() {
		_, err := s.App.BeginBlocker(s.Ctx)
		s.Require().NoError(err)
	})
}

// Modifies sdk config to have stride address prefixes (used for non-keeper tests)
func SetupConfig() {
	app.SetupConfig()
}
