package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type StakingKeeper interface {
	GetValidator(sdk.Context, sdk.ValAddress) (stakingTypes.Validator, bool)
	BondDenom(sdk.Context) string
	Delegate(sdk.Context, sdk.AccAddress, math.Int, stakingTypes.BondStatus, stakingTypes.Validator, bool) (math.LegacyDec, error)
}
