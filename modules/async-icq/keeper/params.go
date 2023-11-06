package keeper

import (
	"github.com/cosmos/ibc-apps/modules/async-icq/v7/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IsHostEnabled retrieves the host enabled boolean from the params.
// True is returned if the host is enabled.
func (k Keeper) IsHostEnabled(ctx sdk.Context) bool {
	return k.GetParams(ctx).HostEnabled
}

// GetAllowQueries retrieves the host enabled query paths from the params
func (k Keeper) GetAllowQueries(ctx sdk.Context) []string {
	return k.GetParams(ctx).AllowQueries
}

// SetParams sets the module parameters.
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) error {
	if err := p.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&p)
	store.Set(types.ParamsKey, bz)
	return nil
}

// GetParams returns the current module parameters.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var p types.Params

	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return p
	}

	k.cdc.MustUnmarshal(bz, &p)
	return p
}
