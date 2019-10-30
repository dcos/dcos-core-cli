package deployment

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/python"
	"github.com/spf13/cobra"
)

func newCmdMarathonDeploymentRollback(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "Remove a deployed application.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return python.InvokePythonCLI(ctx)
		},
	}
	return cmd
}
