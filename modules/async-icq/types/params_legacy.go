/*
NOTE: Usage of x/params to manage parameters is deprecated in favor of x/gov
controlled execution of MsgUpdateParams messages. These types remains solely
for migration purposes and will be removed in a future release.
*/
package types

import (
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	// KeyHostEnabled is the store key for HostEnabled Params
	KeyHostEnabled = []byte("HostEnabled")
	// KeyAllowQueries is the store key for the AllowQueries Params
	KeyAllowQueries = []byte("AllowQueries")
)

// ParamKeyTable type declaration for parameters
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyHostEnabled, &p.HostEnabled, validateEnabled),
		paramtypes.NewParamSetPair(KeyAllowQueries, &p.AllowQueries, validateAllowlist),
	}
}
