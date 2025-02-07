package simapp

import "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v9/testing/simapp/upgrades"

// registerUpgradeHandlers registers all supported upgrade handlers
func (app *SimApp) registerUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler(
		upgrades.V2,
		upgrades.CreateV2UpgradeHandler(app.ModuleManager, app.configurator, app.ParamsKeeper, app.ConsensusParamsKeeper, app.PacketForwardKeeper),
	)
}
