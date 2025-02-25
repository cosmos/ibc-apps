package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

// TransferKeeper defines the expected transfer keeper
type TransferKeeper interface {
	Transfer(ctx context.Context, msg *types.MsgTransfer) (*types.MsgTransferResponse, error)
	DenomPathFromHash(ctx sdk.Context, denom string) (string, error)
	GetTotalEscrowForDenom(ctx sdk.Context, denom string) sdk.Coin
	SetTotalEscrowForDenom(ctx sdk.Context, coin sdk.Coin)

	// Only used for v3 migration
	GetPort(ctx sdk.Context) string
}

// ChannelKeeper defines the expected IBC channel keeper
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel channeltypes.Channel, found bool)
	GetPacketCommitment(ctx sdk.Context, portID, channelID string, sequence uint64) []byte
	GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool)
	LookupModuleByChannel(ctx sdk.Context, portID, channelID string) (string, *capabilitytypes.Capability, error)

	// Only used for v3 migration
	GetAllChannelsWithPortPrefix(ctx sdk.Context, portPrefix string) []channeltypes.IdentifiedChannel
}

// BankKeeper defines the expected bank keeper
type BankKeeper interface {
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error

	// Only used for v3 migration
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}
