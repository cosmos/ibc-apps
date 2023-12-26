package simapp

import (
	appparams "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/testing/simapp/params"

	"github.com/cosmos/cosmos-sdk/std"
)

func MakeEncodingConfig() appparams.EncodingConfig {
	encodingConfig := appparams.MakeTestEncodingConfig()
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}
