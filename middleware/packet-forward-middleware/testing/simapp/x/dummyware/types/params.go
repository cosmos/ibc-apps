package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultFeePercentage is the default value used to extract a fee from all forwarded packets.
var DefaultFeePercentage = sdk.NewDec(0)

// NewParams creates a new parameter configuration for the pfm module.
func NewParams(feePercentage sdk.Dec) Params {
	return Params{}
}

// DefaultParams is the default parameter configuration for the pfm module.
func DefaultParams() Params {
	return NewParams(DefaultFeePercentage)
}

// Validate the pfm module parameters.
func (p Params) Validate() error {
	return nil
}
