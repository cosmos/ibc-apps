package keeper

import (
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
func (k Keeper) GetParams(ctx sdk.Context) (p types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return p
	}

	k.cdc.MustUnmarshal(bz, &p)
	return p
}

// GetFeePercentage retrieves the fee percentage for forwarded packets from the store.
func (k Keeper) GetFeePercentage(ctx sdk.Context) sdk.Dec {
	return k.GetParams(ctx).FeePercentage
}
