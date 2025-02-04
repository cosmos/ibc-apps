package upgrades

import (
	"context"

	appmodule "cosmossdk.io/core/appmodule"
	consensusparamskeeper "cosmossdk.io/x/consensus/keeper"
	paramskeeper "cosmossdk.io/x/params/keeper"
	paramstypes "cosmossdk.io/x/params/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v9/packetforward/keeper"
)

const (
	V2 = "v2"
)

// CreateDefaultUpgradeHandler creates a simple migration upgrade handler.
// func CreateDefaultUpgradeHandler(mm *module.Manager, cfg module.Configurator) upgradetypes.UpgradeHandler {
// 	func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
// 		return mm.RunMigrations(ctx, cfg, fromVM)
// 	}
// }

// We will have to import every one here
func CreateV2UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	paramskeeper paramskeeper.Keeper,
	consensusparamskeeper consensusparamskeeper.Keeper,
	packetforwardkeeper *keeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, fromVM appmodule.VersionMap) (appmodule.VersionMap, error) {
		// Migrate Tendermint consensus parameters from x/params module to a deprecated x/consensus module.
		// The old params module is required to still be imported in your app.go in order to handle this migration.
		baseAppLegacySS := paramskeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())
		if err := baseapp.MigrateParams(sdk.UnwrapSDKContext(ctx), baseAppLegacySS, &consensusparamskeeper.ParamsStore); err != nil {
			return nil, err
		}

		versionMap, err := mm.RunMigrations(ctx, cfg, fromVM)
		if err != nil {
			return nil, err
		}

		return versionMap, err
	}
}
