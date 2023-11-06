package upgrades

import (
	"github.com/cosmos/ibc-apps/modules/async-icq/v7/keeper"
	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v7/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	consensusparamskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const (
	V2 = "v2"
)

// CreateDefaultUpgradeHandler creates a base upgrade handler for the async-icq module.
func CreateDefaultUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// CreateV2UpgradeHandler creates the v2 upgrade handler for the param migration.
func CreateV2UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	paramskeeper paramskeeper.Keeper,
	consensusparamskeeper consensusparamskeeper.Keeper,
	asyncicqkeeper keeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// NOTE: If you already migrated the previous module, you ONLY need to migrate async-icq case now.
		for _, subspace := range paramskeeper.GetSubspaces() {
			subspace := subspace

			var keyTable paramstypes.KeyTable
			if subspace.Name() == icqtypes.ModuleName {
				keyTable = icqtypes.ParamKeyTable()
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
		baseapp.MigrateParams(ctx, baseAppLegacySS, &consensusparamskeeper)

		versionMap, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return nil, err
		}

		return versionMap, err
	}
}
