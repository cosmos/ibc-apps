package test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccAddress returns a random account address
func AccAddress(t *testing.T) sdk.AccAddress {
	t.Helper()
	pk := ed25519.GenPrivKey().PubKey()
	addr := pk.Address()
	return sdk.AccAddress(addr)
}

func AccAddressFromBech32(t *testing.T, addr string) sdk.AccAddress {
	t.Helper()
	a, err := sdk.AccAddressFromBech32(addr)
	require.NoError(t, err)
	return a
}
