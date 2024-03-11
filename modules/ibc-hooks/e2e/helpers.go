package e2e

import (
	"context"
	"strings"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/stretchr/testify/require"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

type WasmCoin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type QueryMsg struct {
	// IBCHooks
	GetCount      *GetCountQuery      `json:"get_count,omitempty"`
	GetTotalFunds *GetTotalFundsQuery `json:"get_total_funds,omitempty"`
}

func SetupContract(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, keyname string, fileLoc string, message string) (codeId, contract string) {
	codeId, err := chain.StoreContract(ctx, keyname, fileLoc)
	if err != nil {
		t.Fatal(err)
	}

	contractAddr, err := chain.InstantiateContract(ctx, keyname, codeId, message, true)
	if err != nil {
		t.Fatal(err)
	}

	return codeId, contractAddr
}

func WasmEncodingConfig() *moduletestutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()
	wasmtypes.RegisterInterfaces(cfg.InterfaceRegistry)
	return &cfg
}

func GetIBCHooksUserAddress(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, channel, uaddr string) string {
	cmd := []string{
		chain.Config().Bin, "query", "ibchooks", "wasm-sender", channel, uaddr,
		"--node", chain.GetRPCAddress(),
		"--chain-id", chain.Config().ChainID,
		"--output", "json",
	}

	// This query does not return a type, just prints the string.
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	address := strings.Replace(string(stdout), "\n", "", -1)
	return address
}

// GetIBCHookTotalFunds
type GetTotalFundsQuery struct {
	// {"get_total_funds":{"addr":"osmo1..."}}
	Addr string `json:"addr"`
}
type GetTotalFundsResponse struct {
	// {"data":{"total_funds":[{"denom":"ibc/04F5F501207C3626A2C14BFEF654D51C2E0B8F7CA578AB8ED272A66FE4E48097","amount":"1"}]}}
	Data *GetTotalFundsObj `json:"data"`
}
type GetTotalFundsObj struct {
	TotalFunds []WasmCoin `json:"total_funds"`
}

func GetIBCHookTotalFunds(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, contract string, uaddr string) GetTotalFundsResponse {
	var res GetTotalFundsResponse
	err := chain.QueryContract(ctx, contract, QueryMsg{GetTotalFunds: &GetTotalFundsQuery{Addr: uaddr}}, &res)
	require.NoError(t, err)
	return res
}

// GetIBCHookCount
type GetCountQuery struct {
	// {"get_total_funds":{"addr":"osmo1..."}}
	Addr string `json:"addr"`
}
type GetCountResponse struct {
	// {"data":{"count":0}}
	Data *GetCountObj `json:"data"`
}
type GetCountObj struct {
	Count int64 `json:"count"`
}

func GetIBCHookCount(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, contract string, uaddr string) GetCountResponse {
	var res GetCountResponse
	err := chain.QueryContract(ctx, contract, QueryMsg{GetCount: &GetCountQuery{Addr: uaddr}}, &res)
	require.NoError(t, err)
	return res
}
