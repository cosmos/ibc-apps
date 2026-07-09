package main

import (
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/cosmos/ibc-apps/modules/async-icq/v8/interchain-query-demo/app"
	"github.com/cosmos/ibc-apps/modules/async-icq/v8/interchain-query-demo/cmd/icqd-demo/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
