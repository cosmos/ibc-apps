package simtests

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm/ibctesting"
	w "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
)

type NoteInstantiate struct {
	BlockMaxGas uint64 `json:"block_max_gas,string"`
}

type VoiceInstantiate struct {
	ProxyCodeId uint64 `json:"proxy_code_id,string"`
	BlockMaxGas uint64 `json:"block_max_gas,string"`
}

type TesterInstantiate struct {
}

type NoteExecute struct {
	Query   *NoteExecuteQuery `json:"query,omitempty"`
	Execute *NoteExecuteMsg   `json:"execute,omitempty"`
}

var (
	NoteQueryActiveChannel = `"active_channel"`
	NoteQueryPair          = `"pair"`
	NoteQueryBlockMaxGas   = `"block_max_gas"`
	NoteQueryRemoteAddress = func(local_address string) string {
		return fmt.Sprintf(`{"remote_address":{"local_address":"%s"}}`, local_address)
	}
)

type NoteExecuteQuery struct {
	Msgs           []w.QueryRequest `json:"msgs"`
	TimeoutSeconds uint64           `json:"timeout_seconds,string"`
	Callback       CallbackRequest  `json:"callback"`
}

type NoteExecuteMsg struct {
	OnBehalfOf     string           `json:"on_behalf_of,omitempty"`
	Msgs           []w.CosmosMsg    `json:"msgs"`
	TimeoutSeconds uint64           `json:"timeout_seconds,string"`
	Callback       *CallbackRequest `json:"callback,omitempty"`
}

type PolytoneMessage struct {
	Query   *PolytoneQuery   `json:"query,omitempty"`
	Execute *PolytoneExecute `json:"execute,omitempty"`
}

type PolytoneQuery struct {
	Msgs []any `json:"msgs"`
}

type PolytoneExecute struct {
	Msgs []any `json:"msgs"`
}

type CallbackRequest struct {
	Receiver string `json:"receiver"`
	Msg      string `json:"msg"`
}

type CallbackMessage struct {
	Initiator    string   `json:"initiator"`
	InitiatorMsg string   `json:"initiator_msg"`
	Result       Callback `json:"result"`
}

type Callback struct {
	Execute    CallbackDataExecute `json:"execute,omitempty"`
	Query      CallbackDataQuery   `json:"query,omitempty"`
	FatalError string              `json:"fatal_error,omitempty"`
}

type CallbackDataQuery struct {
	Ok  [][]byte      `json:"ok,omitempty"`
	Err ErrorResponse `json:"err,omitempty"`
}

type CallbackDataExecute struct {
	Ok  ExecutionResponse `json:"ok,omitempty"`
	Err string            `json:"err,omitempty"`
}

type ExecutionResponse struct {
	ExecutedBy string           `json:"executed_by"`
	Result     []SubMsgResponse `json:"result"`
}

type ErrorResponse struct {
	MessageIndex uint64 `json:"message_index,string"`
	Error        string `json:"error"`
}

type SubMsgResponse struct {
	Events []Event `json:"events"`
	Data   []byte  `json:"data,omitempty"`
}

type Events []Event
type Event struct {
	Type       string          `json:"type"`
	Attributes EventAttributes `json:"attributes"`
}

type EventAttributes []EventAttribute
type EventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Empty struct{}

type TesterQuery struct {
	History      *Empty `json:"history,omitempty"`
	HelloHistory *Empty `json:"hello_history,omitempty"`
}

type HistoryResponse struct {
	History []CallbackMessage `json:"history"`
}

type HelloHistoryResponse struct {
	History []string `json:"history"`
}

func Instantiate(t *testing.T, chain *ibctesting.TestChain, codeId uint64, msg any) sdk.AccAddress {
	instantiate, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	return chain.InstantiateContract(codeId, instantiate)
}

func QueryCallbackHistory(chain *ibctesting.TestChain, tester sdk.AccAddress) []CallbackMessage {
	bytes, err := json.Marshal(TesterQuery{
		History: &Empty{},
	})
	if err != nil {
		panic(err)
	}
	res, err := chain.App.WasmKeeper.QuerySmart(chain.GetContext(), tester, bytes)
	if err != nil {
		panic(err)
	}
	var response HistoryResponse
	json.Unmarshal(res, &response)
	return response.History
}

func QueryHelloHistory(chain *ibctesting.TestChain, tester sdk.AccAddress) []string {
	bytes, err := json.Marshal(TesterQuery{
		HelloHistory: &Empty{},
	})
	if err != nil {
		panic(err)
	}
	res, err := chain.App.WasmKeeper.QuerySmart(chain.GetContext(), tester, bytes)
	if err != nil {
		panic(err)
	}
	var response HelloHistoryResponse
	json.Unmarshal(res, &response)
	return response.History
}

func QueryRemoteAccount(
	chain *ibctesting.TestChain,
	note sdk.AccAddress,
	local_address sdk.AccAddress,
) string {
	query, err := chain.App.WasmKeeper.QuerySmart(
		chain.GetContext(),
		note,
		[]byte(NoteQueryRemoteAddress(local_address.String())),
	)
	if err != nil {
		panic(err)
	}
	return string(query)
}

func QueryBlockMaxGas(
	chain *ibctesting.TestChain,
	note sdk.AccAddress,
) string {
	query, err := chain.App.WasmKeeper.QuerySmart(
		chain.GetContext(),
		note,
		[]byte(NoteQueryBlockMaxGas),
	)
	if err != nil {
		panic(err)
	}
	return string(query)
}

func QueryActiveChannel(
	chain *ibctesting.TestChain,
	note sdk.AccAddress,
) string {
	query, err := chain.App.WasmKeeper.QuerySmart(
		chain.GetContext(),
		note,
		[]byte(NoteQueryActiveChannel),
	)
	if err != nil {
		panic(err)
	}
	return string(query)
}
