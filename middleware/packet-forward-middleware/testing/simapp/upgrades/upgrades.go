package upgrades

import (
	"context"

	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/keeper"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/types"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	consensusparamskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
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
	return func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// NOTE: If you already migrated the previous module, you ONLY need to migrate packetforward case now.
		for _, subspace := range paramskeeper.GetSubspaces() {
			subspace := subspace

			var keyTable paramstypes.KeyTable
			if subspace.Name() == packetforwardtypes.ModuleName {
				keyTable = packetforwardtypes.ParamKeyTable()
			} else {
				continue
			}

			if !subspace.HasKeyTable() {
				subspace.WithKeyTable(keyTable)
			}
		}

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
