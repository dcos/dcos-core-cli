package pkg

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdPackageSearch(ctx api.Context) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Add a package repository to DC/OS",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return invokePythonCLI(ctx)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	return cmd
}
