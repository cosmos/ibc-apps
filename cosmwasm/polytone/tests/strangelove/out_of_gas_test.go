package strangelove

import (
	"testing"

	w "github.com/CosmWasm/wasmvm/types"
	"github.com/stretchr/testify/require"
)

// Tests that the voice module gracefully handles an out-of-gas error
// and returns a callback.
func TestOutOfGas(t *testing.T) {
	suite := NewSuite(t)

	_, _, err := suite.CreateChannel(
		suite.ChainA.Note,
		suite.ChainB.Voice,
		&suite.ChainA,
		&suite.ChainB,
	)
	if err != nil {
		t.Fatal(err)
	}

	testerMsg := `{"hello": { "data": "aGVsbG8K" }}`
	messages := []w.CosmosMsg{}
	for i := 0; i < 300; i++ {
		messages = append(messages, w.CosmosMsg{
			Wasm: &w.WasmMsg{
				Execute: &w.ExecuteMsg{
					ContractAddr: suite.ChainB.Tester,
					Msg:          []byte(testerMsg),
					Funds:        []w.Coin{},
				},
			},
		})
	}

	// first, check that this message works without gas pressure.
	callback, err := suite.RoundtripExecute(suite.ChainA.Note, &suite.ChainA, messages[0:1])
	require.Equal(t, []string{"aGVsbG8K"}, callback.Success, "single message should work")

	// now do 300, this should return an out of gas callback
	callback, err = suite.RoundtripExecute(suite.ChainA.Note, &suite.ChainA, messages)
	if err != nil {
		t.Fatal(err)
	}
	require.Empty(t, callback.Success, "should fail to be executed")
	require.Contains(
		t,
		callback.Error,
		"codespace: sdk, code: 11", // see cosmos-sdk/types/errors/errors.go
		"should run out of gas",
	)
}
