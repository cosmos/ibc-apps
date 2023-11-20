package types

// NewParams creates a new parameter configuration for the module.
func NewParams() Params {
	return Params{}
}

// DefaultParams is the default parameter configuration for the module.
func DefaultParams() Params {
	return NewParams()
}

// Validate the pfm module parameters.
func (p Params) Validate() error {
	return nil
}
