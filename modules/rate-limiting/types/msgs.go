package types

import (
	"regexp"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

const (
	TypeMsgAddRateLimit    = "AddRateLimit"
	TypeMsgUpdateRateLimit = "UpdateRateLimit"
	TypeMsgRemoveRateLimit = "RemoveRateLimit"
	TypeMsgResetRateLimit  = "ResetRateLimit"
)

var (
	_ sdk.Msg = &MsgAddRateLimit{}
	_ sdk.Msg = &MsgUpdateRateLimit{}
	_ sdk.Msg = &MsgRemoveRateLimit{}
	_ sdk.Msg = &MsgResetRateLimit{}

	// Implement legacy interface for ledger support
	_ legacytx.LegacyMsg = &MsgAddRateLimit{}
	_ legacytx.LegacyMsg = &MsgUpdateRateLimit{}
	_ legacytx.LegacyMsg = &MsgRemoveRateLimit{}
	_ legacytx.LegacyMsg = &MsgResetRateLimit{}
)

// ----------------------------------------------
//               MsgAddRateLimit
// ----------------------------------------------

func NewMsgAddRateLimit(denom, channelId string, maxPercentSend sdkmath.Int, maxPercentRecv sdkmath.Int, durationHours uint64) *MsgAddRateLimit {
	return &MsgAddRateLimit{
		Denom:          denom,
		ChannelId:      channelId,
		MaxPercentSend: maxPercentSend,
		MaxPercentRecv: maxPercentRecv,
		DurationHours:  durationHours,
	}
}

func (msg MsgAddRateLimit) Type() string {
	return TypeMsgAddRateLimit
}

func (msg MsgAddRateLimit) Route() string {
	return RouterKey
}

func (msg *MsgAddRateLimit) GetSigners() []sdk.AccAddress {
	staker, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{staker}
}

func (msg *MsgAddRateLimit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddRateLimit) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address (%s)", err)
	}

	if msg.Denom == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid denom (%s)", msg.Denom)
	}

	matched, err := regexp.MatchString(`^channel-\d+$`, msg.ChannelId)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "unable to verify channel-id (%s)", msg.ChannelId)
	}
	if !matched {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"invalid channel-id (%s), must be of the format 'channel-{N}'", msg.ChannelId)
	}

	if msg.MaxPercentSend.GT(sdkmath.NewInt(100)) || msg.MaxPercentSend.LT(sdkmath.ZeroInt()) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"max-percent-send percent must be between 0 and 100 (inclusively), Provided: %v", msg.MaxPercentSend)
	}

	if msg.MaxPercentRecv.GT(sdkmath.NewInt(100)) || msg.MaxPercentRecv.LT(sdkmath.ZeroInt()) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"max-percent-recv percent must be between 0 and 100 (inclusively), Provided: %v", msg.MaxPercentRecv)
	}

	if msg.MaxPercentRecv.IsZero() && msg.MaxPercentSend.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"either the max send or max receive threshold must be greater than 0")
	}

	if msg.DurationHours == 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "duration can not be zero")
	}

	return nil
}

// ----------------------------------------------
//               MsgUpdateRateLimit
// ----------------------------------------------

func NewMsgUpdateRateLimit(denom, channelId string, maxPercentSend sdkmath.Int, maxPercentRecv sdkmath.Int, durationHours uint64) *MsgUpdateRateLimit {
	return &MsgUpdateRateLimit{
		Denom:          denom,
		ChannelId:      channelId,
		MaxPercentSend: maxPercentSend,
		MaxPercentRecv: maxPercentRecv,
		DurationHours:  durationHours,
	}
}

func (msg MsgUpdateRateLimit) Type() string {
	return TypeMsgUpdateRateLimit
}

func (msg MsgUpdateRateLimit) Route() string {
	return RouterKey
}

func (msg *MsgUpdateRateLimit) GetSigners() []sdk.AccAddress {
	staker, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{staker}
}

func (msg *MsgUpdateRateLimit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateRateLimit) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address (%s)", err)
	}

	if msg.Denom == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid denom (%s)", msg.Denom)
	}

	matched, err := regexp.MatchString(`^channel-\d+$`, msg.ChannelId)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "unable to verify channel-id (%s)", msg.ChannelId)
	}
	if !matched {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"invalid channel-id (%s), must be of the format 'channel-{N}'", msg.ChannelId)
	}

	if msg.MaxPercentSend.GT(sdkmath.NewInt(100)) || msg.MaxPercentSend.LT(sdkmath.ZeroInt()) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"max-percent-send percent must be between 0 and 100 (inclusively), Provided: %v", msg.MaxPercentSend)
	}

	if msg.MaxPercentRecv.GT(sdkmath.NewInt(100)) || msg.MaxPercentRecv.LT(sdkmath.ZeroInt()) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"max-percent-recv percent must be between 0 and 100 (inclusively), Provided: %v", msg.MaxPercentRecv)
	}

	if msg.MaxPercentRecv.IsZero() && msg.MaxPercentSend.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"either the max send or max receive threshold must be greater than 0")
	}

	if msg.DurationHours == 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "duration can not be zero")
	}

	return nil
}

// ----------------------------------------------
//               MsgRemoveRateLimit
// ----------------------------------------------

func NewMsgRemoveRateLimit(denom, channelId string) *MsgRemoveRateLimit {
	return &MsgRemoveRateLimit{
		Denom:     denom,
		ChannelId: channelId,
	}
}

func (msg MsgRemoveRateLimit) Type() string {
	return TypeMsgRemoveRateLimit
}

func (msg MsgRemoveRateLimit) Route() string {
	return RouterKey
}

func (msg *MsgRemoveRateLimit) GetSigners() []sdk.AccAddress {
	staker, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{staker}
}

func (msg *MsgRemoveRateLimit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRemoveRateLimit) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address (%s)", err)
	}

	if msg.Denom == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid denom (%s)", msg.Denom)
	}

	matched, err := regexp.MatchString(`^channel-\d+$`, msg.ChannelId)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "unable to verify channel-id (%s)", msg.ChannelId)
	}
	if !matched {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"invalid channel-id (%s), must be of the format 'channel-{N}'", msg.ChannelId)
	}

	return nil
}

// ----------------------------------------------
//               MsgResetRateLimit
// ----------------------------------------------

func NewMsgResetRateLimit(denom, channelId string) *MsgResetRateLimit {
	return &MsgResetRateLimit{
		Denom:     denom,
		ChannelId: channelId,
	}
}

func (msg MsgResetRateLimit) Type() string {
	return TypeMsgResetRateLimit
}

func (msg MsgResetRateLimit) Route() string {
	return RouterKey
}

func (msg *MsgResetRateLimit) GetSigners() []sdk.AccAddress {
	staker, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{staker}
}

func (msg *MsgResetRateLimit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgResetRateLimit) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address (%s)", err)
	}

	if msg.Denom == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid denom (%s)", msg.Denom)
	}

	matched, err := regexp.MatchString(`^channel-\d+$`, msg.ChannelId)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "unable to verify channel-id (%s)", msg.ChannelId)
	}
	if !matched {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"invalid channel-id (%s), must be of the format 'channel-{N}'", msg.ChannelId)
	}

	return nil
}
