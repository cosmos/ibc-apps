package apptesting

import (
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
	tmtypesproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"

	app "github.com/cosmos/ibc-apps/modules/rate-limiting/v7/testing/simapp"
)

var (
	TestChainId = "chain-0"
)

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
	s.Ctx = s.App.BaseApp.NewContext(false, tmtypesproto.Header{Height: 1, ChainID: TestChainId})
	s.QueryHelper = &baseapp.QueryServiceTestHelper{
		GRPCQueryRouter: s.App.GRPCQueryRouter(),
		Ctx:             s.Ctx,
	}
	s.TestAccs = CreateRandomAccounts(3)
}

// Generate random account addresss
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
	_, exists := s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().True(exists)

	s.Ctx = s.Ctx.WithBlockHeight(upgradeHeight)
	s.Require().NotPanics(func() {
		beginBlockRequest := abci.RequestBeginBlock{}
		s.App.BeginBlocker(s.Ctx, beginBlockRequest)
	})
}

// Modifies sdk config to have stride address prefixes (used for non-keeper tests)
func SetupConfig() {
	app.SetupConfig()
}
