package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	// DefaultFeePercentage is the default value used to extract a fee from all forwarded packets.
	DefaultFeePercentage = sdk.NewDec(0)

	// KeyFeePercentage is store's key for FeePercentage Params
	KeyFeePercentage = []byte("FeePercentage")
)

// ParamKeyTable type declaration for parameters.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new parameter configuration for the pfm module.
func NewParams(feePercentage sdk.Dec) Params {
	return Params{
		FeePercentage: feePercentage,
	}
}

// DefaultParams is the default parameter configuration for the pfm module.
func DefaultParams() Params {
	return NewParams(DefaultFeePercentage)
}

// Validate the pfm module parameters.
func (p Params) Validate() error {
	return validateFeePercentage(p.FeePercentage)
}

// ParamSetPairs implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyFeePercentage, p.FeePercentage, validateFeePercentage),
	}
}

// validateFeePercentage asserts that the fee percentage param is a valid sdk.Dec type.
func validateFeePercentage(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("invalid fee percentage. expected not negative, got %d", v.RoundInt64())
	}
	if !(v.LTE(sdk.OneDec())) {
		return fmt.Errorf("invalid fee percentage. expected less than one 1 got %d", v.RoundInt64())
	}

	return nil
}
