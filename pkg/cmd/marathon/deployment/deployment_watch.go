package deployment

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/python"
	"github.com/spf13/cobra"
)

func newCmdMarathonDeploymentWatch(ctx api.Context) *cobra.Command {
	var maxCount int
	var interval int

	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Monitor deployments.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return python.InvokePythonCLI(ctx)
		},
	}

	cmd.Flags().IntVar(&maxCount, "max-count", 0, "Maximum number of entries to fetch and return.")
	// TODO: Python version has no description
	cmd.Flags().IntVar(&interval, "interval", 0, "")

	return cmd
}
