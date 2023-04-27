package types

import (
	"cosmossdk.io/errors"
)

var (
	ErrUnknownDataType    = errors.Register(ModuleName, 1, "unknown data type")
	ErrInvalidChannelFlow = errors.Register(ModuleName, 2, "invalid message sent to channel end")
	ErrInvalidHostPort    = errors.Register(ModuleName, 3, "invalid host port")
	ErrHostDisabled       = errors.Register(ModuleName, 4, "host is disabled")
	ErrInvalidVersion     = errors.Register(ModuleName, 5, "invalid version")
)
