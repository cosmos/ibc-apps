package mocks

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
)

type IBCModuleMock struct {
	porttypes.IBCModule
}

func (m IBCModuleMock) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	return AcknowledgementMock{}
}

func (m IBCModuleMock) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	panic("Mock for OnChanOpenInit panic not yet implemented")
}

func (m IBCModuleMock) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (version string, err error) {
	panic("Mock for OnChanOpenTry panic not yet implemented")
}

func (m IBCModuleMock) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	panic("Mock for OnChanOpenAck panic not yet implemented")
}

func (m IBCModuleMock) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	panic("Mock for OnChanOpenConfirm panic not yet implemented")
}

func (m IBCModuleMock) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	panic("Mock for OnChanCloseInit panic not yet implemented")
}

func (m IBCModuleMock) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	panic("Mock for OnChanCloseConfirm panic not yet implemented")
}

func (m IBCModuleMock) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	panic("Mock for OnAcknowledgementPacket panic not yet implemented")
}

func (m IBCModuleMock) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	panic("Mock for OnTimeoutPacket panic not yet implemented")
}
