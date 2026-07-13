package keeper

import (
	"fmt"

	"github.com/cosmos/ibc-apps/modules/async-icq/v8/types"

	"cosmossdk.io/log/v2"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	porttypes "github.com/cosmos/ibc-go/v11/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v11/modules/core/exported"
)

// Keeper defines the IBC interchain query host keeper
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	ics4Wrapper   porttypes.ICS4Wrapper
	channelKeeper types.ChannelKeeper

	queryRouter *baseapp.GRPCQueryRouter

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

// NewKeeper creates a new interchain query Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec, key storetypes.StoreKey,
	ics4Wrapper porttypes.ICS4Wrapper, channelKeeper types.ChannelKeeper,
	queryRouter *baseapp.GRPCQueryRouter, authority string,
) Keeper {
	return Keeper{
		storeKey:      key,
		cdc:           cdc,
		ics4Wrapper:   ics4Wrapper,
		channelKeeper: channelKeeper,
		queryRouter:   queryRouter,
		authority:     authority,
	}
}

// Logger returns the application logger, scoped to the associated module
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s-%s", ibcexported.ModuleName, types.ModuleName))
}

// SetICS4Wrapper sets the ICS4Wrapper. This function may be used after the
// keeper's creation to set the middleware which is above this module in the
// IBC application stack.
func (k *Keeper) SetICS4Wrapper(wrapper porttypes.ICS4Wrapper) {
	k.ics4Wrapper = wrapper
}

// GetPort returns the portID for the interchain query module. Used in ExportGenesis
func (k Keeper) GetPort(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	return string(store.Get(types.PortKey))
}

// SetPort sets the portID for the interchain query module. Used in InitGenesis
func (k Keeper) SetPort(ctx sdk.Context, portID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.PortKey, []byte(portID))
}

// GetAppVersion calls the ICS4Wrapper GetAppVersion function.
func (k Keeper) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	return k.ics4Wrapper.GetAppVersion(ctx, portID, channelID)
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}
