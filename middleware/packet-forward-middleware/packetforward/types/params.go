package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultFeePercentage is the default value used to extract a fee from all forwarded packets.
var DefaultFeePercentage = sdk.NewDec(0)

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
