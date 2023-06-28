package types

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type PacketMetadata struct {
	Stake *StakeMetadata `json:"stake"`
}

type StakeMetadata struct {
	ValidatorAddress string `json:"validator,omitempty"`
	StakeAmount      string `json:"stake_amount,omitempty"`
}

func (m *StakeMetadata) Validate() error {
	if m.ValidatorAddress == "" {
		return fmt.Errorf("failed to validate forward metadata. validator cannot be empty")
	}
	_, err := sdk.ValAddressFromBech32(m.ValidatorAddress)
	if err != nil {
		return err
	}
	_, ok := math.NewIntFromString(m.StakeAmount)
	if !ok {
		return fmt.Errorf("failed to valdiate stake amount (%s), not a number", m.StakeAmount)
	}
	return nil
}

func (sm *StakeMetadata) ValAddr() sdk.ValAddress {
	val, _ := sdk.ValAddressFromBech32(sm.ValidatorAddress)
	return val
}

func (sm *StakeMetadata) AmountInt() sdk.Int {
	out, _ := math.NewIntFromString(sm.StakeAmount)
	return out
}
