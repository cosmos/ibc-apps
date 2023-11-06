package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"

	abcitypes "github.com/cometbft/cometbft/abci/types"
)

// ModuleCdc references the global interchain queries module codec. Note, the codec
// should ONLY be used in certain instances of tests and for JSON encoding.
//
// The actual codec used for serialization should be provided to interchain queries and
// defined at the application level.
var (
	// ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func SerializeCosmosQuery(reqs []abcitypes.RequestQuery) (bz []byte, err error) {
	q := &CosmosQuery{
		Requests: reqs,
	}
	return ModuleCdc.Marshal(q)
}

func DeserializeCosmosQuery(bz []byte) (reqs []abcitypes.RequestQuery, err error) {
	var q CosmosQuery
	err = ModuleCdc.Unmarshal(bz, &q)
	return q.Requests, err
}

func SerializeCosmosResponse(resps []abcitypes.ResponseQuery) (bz []byte, err error) {
	r := &CosmosResponse{
		Responses: resps,
	}
	return ModuleCdc.Marshal(r)
}

func DeserializeCosmosResponse(bz []byte) (resps []abcitypes.ResponseQuery, err error) {
	var r CosmosResponse
	err = ModuleCdc.Unmarshal(bz, &r)
	return r.Responses, err
}

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	sdk.RegisterLegacyAminoCodec(amino)

	// Register all Amino interfaces and concrete types on the authz Amino codec
	// so that this can later be used to properly serialize MsgGrant and MsgExec
	// instances.
	RegisterLegacyAminoCodec(authzcodec.Amino)
}

// RegisterLegacyAminoCodec registers concrete types on the LegacyAmino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(Params{}, "icq/Params", nil)
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "icq/MsgUpdateParams")
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
