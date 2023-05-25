package helpers

import (
	"testing"

	_ "embed"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/ibc-apps/modules/ibc-hooks/v7/simapp"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/counter/artifacts/counter.wasm
var counterWasm []byte

func SetupEnv(t *testing.T) (*simapp.App, sdk.Context, sdk.AccAddress, *types.BaseAccount) {
	// Setup the environment
	app, ctx, acc := simapp.Setup(t)

	// create the contract because it cannot be mocked
	contractID, _, err := app.ContractKeeper.Create(ctx, acc.GetAddress(), counterWasm, nil)
	require.NoError(t, err)
	contractAddr, _, err := app.ContractKeeper.Instantiate(
		ctx,
		contractID,
		acc.GetAddress(),
		nil,
		[]byte(`{"count": 0}`),
		"test contract",
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, "cosmos14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s4hmalr", contractAddr.String())

	return app, ctx, contractAddr, acc
}
