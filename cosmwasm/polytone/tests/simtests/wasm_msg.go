package simtests

import (
	"encoding/json"

	wasm "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// `WasmMsg::Execute` with the account as the sender.
func (a *Account) WasmExecute(contract *sdk.AccAddress, msg any, funds ...sdk.Coin) sdk.Msg {
	msgstr, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return &wasm.MsgExecuteContract{
		Sender:   a.Address.String(),
		Contract: contract.String(),
		Msg:      wasm.RawContractMessage(msgstr),
		Funds:    funds,
	}
}

// `WasmMsg::Migrate` with the account as the sender.
func (a *Account) WasmMigrate(contract *sdk.AccAddress, msg any, codeId uint64, funds ...sdk.Coin) sdk.Msg {
	msgstr, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return &wasm.MsgMigrateContract{
		Sender:   a.Address.String(),
		Contract: contract.String(),
		Msg:      wasm.RawContractMessage(msgstr),
		CodeID:   codeId,
	}
}

// `WasmMsg::Instantiate` with the account as the sender.
func (a *Account) WasmInstantiate(codeId uint64, msg any, admin *sdk.AccAddress, funds ...sdk.Coin) sdk.Msg {
	msgstr, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	adminstr := ""
	if admin != nil {
		adminstr = admin.String()
	}
	return &wasm.MsgInstantiateContract{
		Sender: a.Address.String(),
		Admin:  adminstr,
		CodeID: codeId,
		Label:  "🌀",
		Msg:    msgstr,
		Funds:  funds,
	}
}
