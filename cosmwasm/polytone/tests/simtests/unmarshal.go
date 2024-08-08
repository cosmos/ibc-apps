package simtests

import (
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm/types"
)

func unmarshalExecute(t *testing.T, data []byte) types.MsgExecuteContractResponse {
	var w types.MsgExecuteContractResponse
	w.Unmarshal(data)
	return w
}

func unmarshalInstantiate(t *testing.T, data []byte) types.MsgInstantiateContractResponse {
	var w types.MsgInstantiateContractResponse
	w.Unmarshal(data)
	return w
}

func unmarshalInstantiate2(t *testing.T, data []byte) types.MsgInstantiateContract2Response {
	var w types.MsgInstantiateContract2Response
	w.Unmarshal(data)
	return w
}
