// from https://github.com/cosmos/ibc-go/blob/a4ef5360b49ad2118e1d68f25f13935162660e0b/modules/apps/transfer/types/coin.go

package keeper

import (
	"fmt"
	"strings"
)

// SenderChainIsSource returns false if the denomination originally came
// from the receiving chain and true otherwise.
func senderChainIsSource(sourcePort, sourceChannel, denom string) bool {
	// This is the prefix that would have been prefixed to the denomination
	// on sender chain IF and only if the token originally came from the
	// receiving chain.

	return !receiverChainIsSource(sourcePort, sourceChannel, denom)
}

// receiverChainIsSource returns true if the denomination originally came
// from the receiving chain and false otherwise.
func receiverChainIsSource(sourcePort, sourceChannel, denom string) bool {
	// The prefix passed in should contain the SourcePort and SourceChannel.
	// If  the receiver chain originally sent the token to the sender chain
	// the denom will have the sender's SourcePort and SourceChannel as the
	// prefix.

	voucherPrefix := getDenomPrefix(sourcePort, sourceChannel)
	return strings.HasPrefix(denom, voucherPrefix)
}

// GetDenomPrefix returns the receiving denomination prefix
func getDenomPrefix(portID, channelID string) string {
	return fmt.Sprintf("%s/%s/", portID, channelID)
}
