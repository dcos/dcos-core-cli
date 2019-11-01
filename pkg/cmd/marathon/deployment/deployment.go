package deployment

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func NewCmdMarathonDeployment(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deployment",
		Short: "Manage deployments.",
	}

	cmd.AddCommand(
		newCmdMarathonDeploymentList(ctx),
		newCmdMarathonDeploymentRollback(ctx),
		newCmdMarathonDeploymentStop(ctx),
		newCmdMarathonDeploymentWatch(ctx),
	)

	return cmd
}
