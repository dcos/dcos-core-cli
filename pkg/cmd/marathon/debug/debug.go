package debug

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func NewCmdMarathonDebug(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Debug app deployments.",
	}

	cmd.AddCommand(
		newCmdMarathonDebugDetails(ctx),
		newCmdMarathonDebugList(ctx),
		newCmdMarathonDebugSummary(ctx),
	)

	return cmd
}

