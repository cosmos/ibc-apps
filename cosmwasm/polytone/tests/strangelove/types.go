package strangelove

import (
	w "github.com/CosmWasm/wasmvm/types"
)

// these types are copied from ../simtests/contract.go. you'll need to
// manually update each one when you make a change. the reason is that
// (1) wasmd ibctesting and interchaintest use different sdk versions
// so they need their own go.mod, (2) i don't know how to use go local
// imports and think it would take more work to learn than to copy
// these files every once and a while.

type NoteInstantiate struct {
}

type VoiceInstantiate struct {
	ProxyCodeId uint64 `json:"proxy_code_id,string"`
	BlockMaxGas uint64 `json:"block_max_gas,string"`
}

type TesterInstantiate struct {
}

type NoteExecute struct {
	Query   *NoteQuery      `json:"query,omitempty"`
	Execute *NoteExecuteMsg `json:"execute,omitempty"`
}

type NoteQuery struct {
	Msgs           []w.CosmosMsg   `json:"msgs"`
	TimeoutSeconds uint64          `json:"timeout_seconds,string"`
	Callback       CallbackRequest `json:"callback"`
}

type NoteExecuteMsg struct {
	Msgs           []w.CosmosMsg    `json:"msgs"`
	TimeoutSeconds uint64           `json:"timeout_seconds,string"`
	Callback       *CallbackRequest `json:"callback,omitempty"`
}

type PolytoneMessage struct {
	Query   *PolytoneQuery   `json:"query,omitempty"`
	Execute *PolytoneExecute `json:"execute,omitempty"`
}

type PolytoneQuery struct {
	Msgs []w.CosmosMsg `json:"msgs"`
}

type PolytoneExecute struct {
	Msgs []w.CosmosMsg `json:"msgs"`
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
	Success []string `json:"success,omitempty"`
	Error   string   `json:"error,omitempty"`
}

type Empty struct{}

type DataWrappedHistoryResponse struct {
	Data HistoryResponse `json:"data"`
}

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
