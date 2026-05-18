package keeper

import (
	v2 "github.com/cosmos/ibc-apps/modules/rate-limiting/v10/migrations/v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper Keeper
}

// NewMigrator creates a new Migrator instance.
func NewMigrator(k Keeper) Migrator {
	return Migrator{keeper: k}
}

// Migrate1to2 widens the PendingSendPacket key's channel-ID segment from
// 16 to 64 bytes so IBC v2 channel IDs fit.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v2.Migrate(ctx, m.keeper.storeService)
}
