package keeper

import (
	"fmt"

	"github.com/cosmos/ibc-apps/modules/rate-limiting/v7/types"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/cometbft/cometbft/libs/log"
)

type (
	Keeper struct {
		storeKey   storetypes.StoreKey
		cdc        codec.BinaryCodec
		paramstore paramtypes.Subspace
		authority  string

		bankKeeper    types.BankKeeper
		channelKeeper types.ChannelKeeper
		ics4Wrapper   types.ICS4Wrapper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	ps paramtypes.Subspace,
	authority string,
	bankKeeper types.BankKeeper,
	channelKeeper types.ChannelKeeper,
	ics4Wrapper types.ICS4Wrapper,
) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      key,
		paramstore:    ps,
		authority:     authority,
		bankKeeper:    bankKeeper,
		channelKeeper: channelKeeper,
		ics4Wrapper:   ics4Wrapper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}
