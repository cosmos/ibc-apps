package types

import "fmt"

const (
	// ModuleName defines the 29-fee name
	ModuleName = "packetfowardmiddleware"

	// StoreKey is the store key string for IBC transfer
	StoreKey = ModuleName

	// QuerierRoute is the querier route for IBC transfer
	QuerierRoute = ModuleName
)

type (
	NonrefundableKey           struct{}
	DisableDenomCompositionKey struct{}
	ProcessedKey               struct{}
)

func RefundPacketKey(channelID, portID string, sequence uint64) []byte {
	return []byte(fmt.Sprintf("%s/%s/%d", channelID, portID, sequence))
}
