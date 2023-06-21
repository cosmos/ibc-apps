package types

import "fmt"

// NewGenesisState creates a 29-fee GenesisState instance.
func NewGenesisState(params Params, inFlightPackets map[string]InFlightPacket) *GenesisState {
	return &GenesisState{
		Params:          params,
		InFlightPackets: inFlightPackets,
	}
}

// DefaultGenesisState returns a GenesisState with "transfer" as the default PortID.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:          DefaultParams(),
		InFlightPackets: make(map[string]InFlightPacket),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("failed to validate genesis params: %w", err)
	}

	if gs.InFlightPackets == nil {
		return fmt.Errorf("in flight packets not initialized in genesis")
	}

	return nil
}
