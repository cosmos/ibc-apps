package types

import (
	icqtypes "github.com/cosmos/ibc-go/v3/modules/apps/icq/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "interquery"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_interquery"

	// Version defines the current version the IBC module supports
	Version = icqtypes.Version

	// PortID is the default port id that module binds to
	PortID = "interquery"
)

var (
	// PortKey defines the key to store the port ID in store
	PortKey = KeyPrefix("interquery-port-")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
