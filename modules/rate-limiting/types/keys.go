package types

import (
	"encoding/binary"

	errorsmod "cosmossdk.io/errors"
)

const (
	ModuleName = "ratelimit"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

var (
	PathKeyPrefix             = KeyPrefix("path")
	RateLimitKeyPrefix        = KeyPrefix("rate-limit")
	PendingSendPacketPrefix   = KeyPrefix("pending-send-packet")
	DenomBlacklistKeyPrefix   = KeyPrefix("denom-blacklist")
	AddressWhitelistKeyPrefix = KeyPrefix("address-blacklist")
	HourEpochKey              = KeyPrefix("hour-epoch")

	PendingSendPacketChannelLength int = 64
)

// Get the rate limit byte key built from the denom and channelId
func GetRateLimitItemKey(denom string, channelId string) []byte {
	return append(KeyPrefix(denom), KeyPrefix(channelId)...)
}

// Get the pending send packet key from the channel ID and sequence number
// The channel ID must be fixed length to allow for extracting the underlying
// values from a key
func GetPendingSendPacketKey(channelId string, sequenceNumber uint64) ([]byte, error) {
	if len(channelId) > PendingSendPacketChannelLength {
		return nil, errorsmod.Wrapf(ErrInvalidChannelId, "channel %s with length %d is greater than the allowed length %d", channelId, len(channelId), PendingSendPacketChannelLength)
	}
	channelIdBz := make([]byte, PendingSendPacketChannelLength)
	copy(channelIdBz, channelId)

	sequenceNumberBz := make([]byte, 8)
	binary.BigEndian.PutUint64(sequenceNumberBz, sequenceNumber)

	return append(channelIdBz, sequenceNumberBz...), nil
}

// Get the whitelist path key from a sender and receiver address
func GetAddressWhitelistKey(sender, receiver string) []byte {
	return append(KeyPrefix(sender), KeyPrefix(receiver)...)
}
