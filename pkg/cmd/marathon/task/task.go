package task

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func NewCmdMarathonTask(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage tasks.",
	}

	cmd.AddCommand(
		newCmdMarathonTaskKill(ctx),
		newCmdMarathonTaskList(ctx),
		newCmdMarathonTaskShow(ctx),
		newCmdMarathonTaskStop(ctx),
	)

	return cmd
}
