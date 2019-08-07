package pkg

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdPackageList(ctx api.Context) *cobra.Command {
	var cliOnly, jsonOutput bool
	var appID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Print a list of the installed DC/OS packages",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return invokePythonCLI(ctx)
		},
	}
	cmd.Flags().StringVar(&appID, "app-id", "", "The application ID")
	cmd.Flags().BoolVar(&cliOnly, "cli", false, "Command line only")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	return cmd
}
