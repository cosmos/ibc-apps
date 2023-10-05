package main

import (
	"os"

	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/testing/simapp"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/testing/simapp/simd/cmd"

	"cosmossdk.io/log"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
)

func main() {
	rootCmd, _ := cmd.NewRootCmd()

	if err := svrcmd.Execute(rootCmd, "", simapp.DefaultNodeHome); err != nil {
		log.NewLogger(rootCmd.OutOrStderr()).Error("failure when running app", "err", err)
		os.Exit(1)
	}
}
