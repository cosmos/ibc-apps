package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TransferKeeper defines the expected transfer keeper
// type TransferKeeper interface {
// 	Transfer(ctx context.Context, msg *types.MsgTransfer) (*types.MsgTransferResponse, error)
// 	DenomPathFromHash(ctx sdk.Context, denom string) (string, error)
// }

// // ChannelKeeper defines the expected IBC channel keeper
// type ChannelKeeper interface {
// 	GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel channeltypes.Channel, found bool)
// 	GetPacketCommitment(ctx sdk.Context, portID, channelID string, sequence uint64) []byte
// 	GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool)
// 	LookupModuleByChannel(ctx sdk.Context, portID, channelID string) (string, *capabilitytypes.Capability, error)
// }

// // DistributionKeeper defines the expected distribution keeper
// type DistributionKeeper interface {
// 	FundCommunityPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error
// }

// // BankKeeper defines the expected bank keeper
// type BankKeeper interface {
// 	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
// 	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
// 	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
// }

type StakingKeeper interface {
	GetValidator(sdk.Context, sdk.ValAddress) (stakingTypes.Validator, bool)
	BondDenom(sdk.Context) string
	Delegate(sdk.Context, sdk.AccAddress, math.Int, stakingTypes.BondStatus, stakingTypes.Validator, bool) (math.LegacyDec, error)
}
