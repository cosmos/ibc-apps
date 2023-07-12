package types

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		Axelar: nil,
	}
}

// ValidateBasic performs basic validation on parameters.
func (p Params) ValidateBasic() error {
	return nil
}
