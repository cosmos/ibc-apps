package v3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/types"
)

// Migrate migrates the x/packetforward module state from the consensus version
// 2 to version 3
func Migrate(
	ctx sdk.Context,
	bankKeeper types.BankKeeper,
	channelKeeper types.ChannelKeeper,
	transferKeeper types.TransferKeeper,
) error {

	return nil
}
