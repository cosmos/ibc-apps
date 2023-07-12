package types

var (
	ParamsKey = []byte{0x01}
)

const (
	ModuleName     = "ibchooks"
	StoreKey       = "hooks-for-ibc" // not using the module name because of collisions with key "ibc"
	IBCCallbackKey = "ibc_callback"
	SenderPrefix   = "ibc-wasm-hook-intermediary"
)
