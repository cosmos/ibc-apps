package cli

import (
	"fmt"

	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/testing/simapp/x/dummyware/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
)

// GetQueryCmd returns the query commands for dummyware
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the " + types.ModuleName + " module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	queryCmd.AddCommand(
		GetCmdParams(),
	)

	return queryCmd
}

// GetCmdParams returns the command handler for dummyware parameter querying.
func GetCmdParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "params",
		Short:   "Query the current dummyware parameters",
		Long:    "Query the current dummyware parameters",
		Args:    cobra.NoArgs,
		Example: fmt.Sprintf("%s query dummyware params", version.AppName),
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// NewTxCmd returns the transaction commands for dummyware
func NewTxCmd() *cobra.Command {
	return nil
}
