package keeper

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/x/group/errors"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/ibc-apps/modules/ibc-hooks/v7/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	Keeper struct {
		cdc       codec.BinaryCodec
		storeKey  storetypes.StoreKey
		authority string
	}
)

// NewKeeper returns a new instance of the x/ibchooks keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	authority string,
) Keeper {
	return Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		authority: authority,
	}
}

// Logger returns a logger for the x/tokenfactory module
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func GetPacketKey(channel string, packetSequence uint64) []byte {
	return []byte(fmt.Sprintf("%s::%d", channel, packetSequence))
}

// StorePacketCallback stores which contract will be listening for the ack or timeout of a packet
func (k Keeper) StorePacketCallback(ctx sdk.Context, channel string, packetSequence uint64, contract string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(GetPacketKey(channel, packetSequence), []byte(contract))
}

// GetPacketCallback returns the bech32 addr of the contract that is expecting a callback from a packet
func (k Keeper) GetPacketCallback(ctx sdk.Context, channel string, packetSequence uint64) string {
	store := ctx.KVStore(k.storeKey)
	return string(store.Get(GetPacketKey(channel, packetSequence)))
}

// DeletePacketCallback deletes the callback from storage once it has been processed
func (k Keeper) DeletePacketCallback(ctx sdk.Context, channel string, packetSequence uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetPacketKey(channel, packetSequence))
}

func (k Keeper) DeriveIntermediateSender(ctx sdk.Context, channel, originalSender, bech32Prefix string, wasm map[string]interface{}) (string, error) {
	// If we have trusted Axelar config available on params, then we can use the source_address
	// value in the Axelar payload to set the sender.
	axelarParams := k.GetParams(ctx).Axelar
	if axelarSender, err := deriveAxelarSender(axelarParams, channel, originalSender, wasm); err == nil {
		return createAddress(axelarSender, bech32Prefix)
	} else {
		return DeriveDefaultIntermediateSender(channel, originalSender, bech32Prefix)
	}
}

func DeriveDefaultIntermediateSender(channel, originalSender, bech32Prefix string) (string, error) {

	senderStr := fmt.Sprintf("%s/%s", channel, originalSender)
	senderHash32 := address.Hash(types.SenderPrefix, []byte(senderStr))

	return createAddress(senderHash32, bech32Prefix)

}

func createAddress(bytes []byte, bech32Prefix string) (string, error) {
	sender := sdk.AccAddress(bytes[:])
	return sdk.Bech32ifyAddressBytes(bech32Prefix, sender)

}

func deriveAxelarSender(params *types.Axelar, channel, originalSender string, wasm map[string]interface{}) ([]byte, error) {
	if params != nil &&
		params.ChannelId == channel &&
		params.GmpAccount == originalSender {

		sourceAddress, ok := wasm["source_address"]

		if !ok {
			return nil, errors.ErrEmpty
		}
		sourceAddressString, ok := sourceAddress.(string)
		sourceAddressString = strings.TrimLeft(sourceAddressString, "0x")

		if !ok {
			return nil, errors.ErrInvalid
		}

		sourceAddressBytes, err := hex.DecodeString(sourceAddressString)

		if err != nil {
			return nil, errors.ErrInvalid
		}
		return sourceAddressBytes, nil
	}
	return nil, errors.ErrInvalid
}

// SetParams sets the ibc-hooks module parameters.
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) error {
	if err := p.ValidateBasic(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&p)
	store.Set(types.ParamsKey, bz)

	return nil
}

// GetParams returns the current ibc-hooks module parameters.
func (k Keeper) GetParams(ctx sdk.Context) (p types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return p
	}

	k.cdc.MustUnmarshal(bz, &p)
	return p
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{Params: k.GetParams(ctx)}
}
