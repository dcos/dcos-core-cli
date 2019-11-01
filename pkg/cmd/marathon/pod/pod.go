package pod

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func NewCmdMarathonPod(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pod",
		Short: "Manage pods.",
	}

	cmd.AddCommand(
		newCmdMarathonPodAdd(ctx),
		newCmdMarathonPodKill(ctx),
		newCmdMarathonPodList(ctx),
		newCmdMarathonPodRemove(ctx),
		newCmdMarathonPodShow(ctx),
		newCmdMarathonPodUpdate(ctx),
	)

	return cmd
}
