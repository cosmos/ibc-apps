package keeper

import (
	"github.com/cosmos/ibc-apps/middleware/staking-middleware/v7/staking/types"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cometbft/cometbft/libs/log"

	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
)

// Keeper defines the packet forward middleware keeper
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	StakingKeeper types.StakingKeeper
	ICS4Wrapper   porttypes.ICS4Wrapper
}

// NewKeeper creates a new forward Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	stakingKeeper types.StakingKeeper,
	ics4wrapper porttypes.ICS4Wrapper,
) *Keeper {

	return &Keeper{
		cdc:           cdc,
		storeKey:      key,
		StakingKeeper: stakingKeeper,
		ICS4Wrapper:   ics4wrapper,
	}
}

// Logger returns a module-specific logger.
func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+ibcexported.ModuleName+"-"+types.ModuleName)
}
