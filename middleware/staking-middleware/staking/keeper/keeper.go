package keeper

import (
	"github.com/cosmos/ibc-apps/middleware/staking-middleware/v7/staking/types"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cometbft/cometbft/libs/log"

	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
)

// Keeper defines the packet forward middleware keeper
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	StakingKeeper types.StakingKeeper
}

// NewKeeper creates a new forward Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	stakingKeeper types.StakingKeeper,
) *Keeper {

	return &Keeper{
		cdc:           cdc,
		storeKey:      key,
		StakingKeeper: stakingKeeper,
	}
}

// Logger returns a module-specific logger.
func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+ibcexported.ModuleName+"-"+types.ModuleName)
}
