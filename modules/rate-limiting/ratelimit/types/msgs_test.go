package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/ibc-rate-limiting/ratelimit/types"
	"github.com/Stride-Labs/ibc-rate-limiting/testing/simapp/apptesting"
)

// ----------------------------------------------
//               MsgAddRateLimit
// ----------------------------------------------

func TestMsgAddRateLimit(t *testing.T) {
	apptesting.SetupConfig()

	validAuthority := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	validDenom := "denom"
	validChannelId := "channel-0"
	validMaxPercentSend := sdkmath.NewInt(10)
	validMaxPercentRecv := sdkmath.NewInt(10)
	validDurationHours := uint64(60)

	testCases := []struct {
		name string
		msg  types.MsgAddRateLimit
		err  string
	}{
		{
			name: "successful proposal",
			msg: types.MsgAddRateLimit{
				Authority:      validAuthority,
				Denom:          validDenom,
				ChannelId:      validChannelId,
				MaxPercentSend: validMaxPercentSend,
				MaxPercentRecv: validMaxPercentRecv,
				DurationHours:  validDurationHours,
			},
		},
		{
			name: "invalid authority",
			msg: types.MsgAddRateLimit{
				Authority:      "invalid_address",
				Denom:          validDenom,
				ChannelId:      validChannelId,
				MaxPercentSend: validMaxPercentSend,
				MaxPercentRecv: validMaxPercentRecv,
				DurationHours:  validDurationHours,
			},
			err: "invalid authority",
		},
		{
			name: "invalid denom",
			msg: types.MsgAddRateLimit{
				Authority:      validAuthority,
				Denom:          "",
				ChannelId:      validChannelId,
				MaxPercentSend: validMaxPercentSend,
				MaxPercentRecv: validMaxPercentRecv,
				DurationHours:  validDurationHours,
			},
			err: "invalid denom",
		},
		{
			name: "invalid channel-id",
			msg: types.MsgAddRateLimit{
				Authority:      validAuthority,
				Denom:          validDenom,
				ChannelId:      "channel-",
				MaxPercentSend: validMaxPercentSend,
				MaxPercentRecv: validMaxPercentRecv,
				DurationHours:  validDurationHours,
			},
			err: "invalid channel-id",
		},
		{
			name: "invalid send percent (lt 0)",
			msg: types.MsgAddRateLimit{
				Authority:      validAuthority,
				Denom:          validDenom,
				ChannelId:      validChannelId,
				MaxPercentSend: sdkmath.NewInt(-1),
				MaxPercentRecv: validMaxPercentRecv,
				DurationHours:  validDurationHours,
			},
			err: "percent must be between 0 and 100",
		},
		{
			name: "invalid send percent (gt 100)",
			msg: types.MsgAddRateLimit{
				Authority:      validAuthority,
				Denom:          validDenom,
				ChannelId:      validChannelId,
				MaxPercentSend: sdkmath.NewInt(101),
				MaxPercentRecv: validMaxPercentRecv,
				DurationHours:  validDurationHours,
			},
			err: "percent must be between 0 and 100",
		},
		{
			name: "invalid receive percent (lt 0)",
			msg: types.MsgAddRateLimit{
				Authority:      validAuthority,
				Denom:          validDenom,
				ChannelId:      validChannelId,
				MaxPercentSend: validMaxPercentSend,
				MaxPercentRecv: sdkmath.NewInt(-1),
				DurationHours:  validDurationHours,
			},
			err: "percent must be between 0 and 100",
		},
		{
			name: "invalid receive percent (gt 100)",
			msg: types.MsgAddRateLimit{
				Authority:      validAuthority,
				Denom:          validDenom,
				ChannelId:      validChannelId,
				MaxPercentSend: validMaxPercentSend,
				MaxPercentRecv: sdkmath.NewInt(101),
				DurationHours:  validDurationHours,
			},
			err: "percent must be between 0 and 100",
		},
		{
			name: "invalid send and receive percent",
			msg: types.MsgAddRateLimit{
				Authority:      validAuthority,
				Denom:          validDenom,
				ChannelId:      validChannelId,
				MaxPercentSend: sdkmath.ZeroInt(),
				MaxPercentRecv: sdkmath.ZeroInt(),
				DurationHours:  validDurationHours,
			},
			err: "either the max send or max receive threshold must be greater than 0",
		},
		{
			name: "invalid duration",
			msg: types.MsgAddRateLimit{
				Authority:      validAuthority,
				Denom:          validDenom,
				ChannelId:      validChannelId,
				MaxPercentSend: validMaxPercentSend,
				MaxPercentRecv: validMaxPercentRecv,
				DurationHours:  0,
			},
			err: "duration can not be zero",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == "" {
				require.NoError(t, tc.msg.ValidateBasic(), "test: %v", tc.name)
				require.Equal(t, tc.msg.Denom, validDenom, "denom")
				require.Equal(t, tc.msg.ChannelId, validChannelId, "channel-id")
				require.Equal(t, tc.msg.MaxPercentSend, validMaxPercentSend, "maxPercentSend")
				require.Equal(t, tc.msg.MaxPercentRecv, validMaxPercentRecv, "maxPercentRecv")
				require.Equal(t, tc.msg.DurationHours, validDurationHours, "durationHours")

				require.Equal(t, tc.msg.Type(), types.TypeMsgAddRateLimit, "type")
				require.Equal(t, tc.msg.Route(), types.ModuleName, "route")
			} else {
				require.ErrorContains(t, tc.msg.ValidateBasic(), tc.err, "test: %v", tc.name)
			}
		})
	}
}

// ----------------------------------------------
//               MsgUpdateRateLimit
// ----------------------------------------------

func TestMsgUpdateRateLimit(t *testing.T) {
	apptesting.SetupConfig()

	validAuthority := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	validDenom := "denom"
	validChannelId := "channel-0"
	validMaxPercentSend := sdkmath.NewInt(10)
	validMaxPercentRecv := sdkmath.NewInt(10)
	validDurationHours := uint64(60)

	testCases := []struct {
		name string
		msg  types.MsgUpdateRateLimit
		err  string
	}{
		{
			name: "successful proposal",
			msg: types.MsgUpdateRateLimit{
				Authority:      validAuthority,
				Denom:          validDenom,
				ChannelId:      validChannelId,
				MaxPercentSend: validMaxPercentSend,
				MaxPercentRecv: validMaxPercentRecv,
				DurationHours:  validDurationHours,
			},
		},
		{
			name: "invalid authority",
			msg: types.MsgUpdateRateLimit{
				Authority:      "invalid_address",
				Denom:          validDenom,
				ChannelId:      validChannelId,
				MaxPercentSend: validMaxPercentSend,
				MaxPercentRecv: validMaxPercentRecv,
				DurationHours:  validDurationHours,
			},
			err: "invalid authority",
		},
		{
			name: "invalid denom",
			msg: types.MsgUpdateRateLimit{
				Authority:      validAuthority,
				Denom:          "",
				ChannelId:      validChannelId,
				MaxPercentSend: validMaxPercentSend,
				MaxPercentRecv: validMaxPercentRecv,
				DurationHours:  validDurationHours,
			},
			err: "invalid denom",
		},
		{
			name: "invalid channel-id",
			msg: types.MsgUpdateRateLimit{
				Authority:      validAuthority,
				Denom:          validDenom,
				ChannelId:      "channel-",
				MaxPercentSend: validMaxPercentSend,
				MaxPercentRecv: validMaxPercentRecv,
				DurationHours:  validDurationHours,
			},
			err: "invalid channel-id",
		},
		{
			name: "invalid send percent (lt 0)",
			msg: types.MsgUpdateRateLimit{
				Authority:      validAuthority,
				Denom:          validDenom,
				ChannelId:      validChannelId,
				MaxPercentSend: sdkmath.NewInt(-1),
				MaxPercentRecv: validMaxPercentRecv,
				DurationHours:  validDurationHours,
			},
			err: "percent must be between 0 and 100",
		},
		{
			name: "invalid send percent (gt 100)",
			msg: types.MsgUpdateRateLimit{
				Authority:      validAuthority,
				Denom:          validDenom,
				ChannelId:      validChannelId,
				MaxPercentSend: sdkmath.NewInt(101),
				MaxPercentRecv: validMaxPercentRecv,
				DurationHours:  validDurationHours,
			},
			err: "percent must be between 0 and 100",
		},
		{
			name: "invalid receive percent (lt 0)",
			msg: types.MsgUpdateRateLimit{
				Authority:      validAuthority,
				Denom:          validDenom,
				ChannelId:      validChannelId,
				MaxPercentSend: validMaxPercentSend,
				MaxPercentRecv: sdkmath.NewInt(-1),
				DurationHours:  validDurationHours,
			},
			err: "percent must be between 0 and 100",
		},
		{
			name: "invalid receive percent (gt 100)",
			msg: types.MsgUpdateRateLimit{
				Authority:      validAuthority,
				Denom:          validDenom,
				ChannelId:      validChannelId,
				MaxPercentSend: validMaxPercentSend,
				MaxPercentRecv: sdkmath.NewInt(101),
				DurationHours:  validDurationHours,
			},
			err: "percent must be between 0 and 100",
		},
		{
			name: "invalid send and receive percent",
			msg: types.MsgUpdateRateLimit{
				Authority:      validAuthority,
				Denom:          validDenom,
				ChannelId:      validChannelId,
				MaxPercentSend: sdkmath.ZeroInt(),
				MaxPercentRecv: sdkmath.ZeroInt(),
				DurationHours:  validDurationHours,
			},
			err: "either the max send or max receive threshold must be greater than 0",
		},
		{
			name: "invalid duration",
			msg: types.MsgUpdateRateLimit{
				Authority:      validAuthority,
				Denom:          validDenom,
				ChannelId:      validChannelId,
				MaxPercentSend: validMaxPercentSend,
				MaxPercentRecv: validMaxPercentRecv,
				DurationHours:  0,
			},
			err: "duration can not be zero",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == "" {
				require.NoError(t, tc.msg.ValidateBasic(), "test: %v", tc.name)
				require.Equal(t, tc.msg.Denom, validDenom, "denom")
				require.Equal(t, tc.msg.ChannelId, validChannelId, "channel-id")
				require.Equal(t, tc.msg.MaxPercentSend, validMaxPercentSend, "maxPercentSend")
				require.Equal(t, tc.msg.MaxPercentRecv, validMaxPercentRecv, "maxPercentRecv")
				require.Equal(t, tc.msg.DurationHours, validDurationHours, "durationHours")

				require.Equal(t, tc.msg.Type(), types.TypeMsgUpdateRateLimit, "type")
				require.Equal(t, tc.msg.Route(), types.ModuleName, "route")
			} else {
				require.ErrorContains(t, tc.msg.ValidateBasic(), tc.err, "test: %v", tc.name)
			}
		})
	}
}

// ----------------------------------------------
//               MsgRemoveRateLimit
// ----------------------------------------------

func TestMsgRemoveRateLimit(t *testing.T) {
	apptesting.SetupConfig()

	validAuthority := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	validDenom := "denom"
	validChannelId := "channel-0"

	testCases := []struct {
		name string
		msg  types.MsgRemoveRateLimit
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgRemoveRateLimit{
				Authority: validAuthority,
				Denom:     validDenom,
				ChannelId: validChannelId,
			},
		},
		{
			name: "invalid authority",
			msg: types.MsgRemoveRateLimit{
				Authority: "invalid_address",
				Denom:     validDenom,
				ChannelId: validChannelId,
			},
			err: "invalid authority",
		},
		{
			name: "invalid denom",
			msg: types.MsgRemoveRateLimit{
				Authority: validAuthority,
				Denom:     "",
				ChannelId: validChannelId,
			},
			err: "invalid denom",
		},
		{
			name: "invalid channel-id",
			msg: types.MsgRemoveRateLimit{
				Authority: validAuthority,
				Denom:     validDenom,
				ChannelId: "chan-1",
			},
			err: "invalid channel-id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == "" {
				require.NoError(t, tc.msg.ValidateBasic(), "test: %v", tc.name)
				require.Equal(t, tc.msg.Denom, validDenom, "denom")
				require.Equal(t, tc.msg.ChannelId, validChannelId, "channelId")

				require.Equal(t, tc.msg.Type(), types.TypeMsgRemoveRateLimit, "type")
				require.Equal(t, tc.msg.Route(), types.ModuleName, "route")
			} else {
				require.ErrorContains(t, tc.msg.ValidateBasic(), tc.err, "test: %v", tc.name)
			}
		})
	}
}

// ----------------------------------------------
//               MsgResetRateLimit
// ----------------------------------------------

func TestMsgResetRateLimit(t *testing.T) {
	apptesting.SetupConfig()

	validAuthority := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	validDenom := "denom"
	validChannelId := "channel-0"

	testCases := []struct {
		name string
		msg  types.MsgResetRateLimit
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgResetRateLimit{
				Authority: validAuthority,
				Denom:     validDenom,
				ChannelId: validChannelId,
			},
		},
		{
			name: "invalid authority",
			msg: types.MsgResetRateLimit{
				Authority: "invalid_address",
				Denom:     validDenom,
				ChannelId: validChannelId,
			},
			err: "invalid authority",
		},
		{
			name: "invalid denom",
			msg: types.MsgResetRateLimit{
				Authority: validAuthority,
				Denom:     "",
				ChannelId: validChannelId,
			},
			err: "invalid denom",
		},
		{
			name: "invalid channel-id",
			msg: types.MsgResetRateLimit{
				Authority: validAuthority,
				Denom:     validDenom,
				ChannelId: "chan-1",
			},
			err: "invalid channel-id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == "" {
				require.NoError(t, tc.msg.ValidateBasic(), "test: %v", tc.name)
				require.Equal(t, tc.msg.Denom, validDenom, "denom")
				require.Equal(t, tc.msg.ChannelId, validChannelId, "channelId")

				require.Equal(t, tc.msg.Type(), types.TypeMsgResetRateLimit, "type")
				require.Equal(t, tc.msg.Route(), types.ModuleName, "route")
			} else {
				require.ErrorContains(t, tc.msg.ValidateBasic(), tc.err, "test: %v", tc.name)
			}
		})
	}
}
