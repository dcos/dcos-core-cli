package leader

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func NewCmdMarathonLeader(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "leader",
		Short: "Manage leader.",
	}

	cmd.AddCommand(
		newCmdMarathonLeaderDelete(ctx),
		newCmdMarathonLeaderShow(ctx),
	)

	return cmd
}

