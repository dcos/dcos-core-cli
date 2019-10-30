package deployment

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/python"
	"github.com/spf13/cobra"
)

func newCmdMarathonDeploymentList(ctx api.Context) *cobra.Command {
	var json bool
	var quiet bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Print a list of currently deployed applications.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return python.InvokePythonCLI(ctx)
		},
	}

	cmd.Flags().BoolVar(&json, "json", false, "Print JSON-formatted data.")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Display IDs only for list.")

	return cmd
}
