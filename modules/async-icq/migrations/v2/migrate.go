package v2

import (
	"fmt"

	"github.com/cosmos/ibc-apps/modules/async-icq/v7/exported"
	"github.com/cosmos/ibc-apps/modules/async-icq/v7/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Migrate migrates the x/async-icq module state from the consensus version 1 to
// version 2. Specifically, it takes the parameters that are currently stored
// and managed by the x/params modules and stores them directly into the x/async-icq
// module state.
func Migrate(
	ctx sdk.Context,
	store sdk.KVStore,
	legacySubspace exported.Subspace,
	cdc codec.BinaryCodec,
) error {
	var currParams types.Params
	legacySubspace.GetParamSet(ctx, &currParams)

	if err := currParams.Validate(); err != nil {
		return err
	}

	bz := cdc.MustMarshal(&currParams)
	store.Set(types.ParamsKey, bz)

	return validate(store, cdc, currParams)
}

func validate(store sdk.KVStore, cdc codec.BinaryCodec, currParams types.Params) error {
	var res types.Params
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return fmt.Errorf("expected params at key %s but not found", types.ParamsKey)
	}

	if err := cdc.Unmarshal(bz, &res); err != nil {
		return err
	}

	if currParams.HostEnabled != res.HostEnabled {
		return fmt.Errorf("expected %+v but got %+v", currParams, res)
	}

	if len(currParams.AllowQueries) != len(res.AllowQueries) {
		return fmt.Errorf("expected %+v but got %+v", currParams, res)
	}

	return nil
}
