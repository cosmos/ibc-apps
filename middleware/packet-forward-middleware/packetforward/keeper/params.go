package keeper

import (
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v6/packetforward/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetFeePercentage retrieves the fee percentage for forwarded packets from the store.
func (k Keeper) GetFeePercentage(ctx sdk.Context) sdk.Dec {
	var res sdk.Dec
	k.paramSpace.Get(ctx, types.KeyFeePercentage, &res)
	return res
}

// GetParams returns the total set of pfm parameters.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(k.GetFeePercentage(ctx))
}

// SetParams sets the total set of pfm parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
