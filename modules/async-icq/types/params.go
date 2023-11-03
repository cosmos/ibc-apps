package types

import (
	"fmt"
	"strings"
)

const (
	// DefaultHostEnabled is the default value for the host param (set to true)
	DefaultHostEnabled = true
)

// NewParams creates a new parameter configuration
func NewParams(enableHost bool, allowQueries []string) Params {
	return Params{
		HostEnabled:  enableHost,
		AllowQueries: allowQueries,
	}
}

// DefaultParams is the default parameter configuration
func DefaultParams() Params {
	return NewParams(DefaultHostEnabled, nil)
}

// Validate validates all parameters
func (p Params) Validate() error {
	if err := validateEnabled(p.HostEnabled); err != nil {
		return err
	}
	return validateAllowlist(p.AllowQueries)
}

func validateEnabled(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateAllowlist(i interface{}) error {
	allowQueries, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, path := range allowQueries {
		if strings.TrimSpace(path) == "" {
			return fmt.Errorf("parameter must not contain empty strings: %s", allowQueries)
		}
	}

	return nil
}
