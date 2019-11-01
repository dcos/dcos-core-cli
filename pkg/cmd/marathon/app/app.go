package app

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func NewCmdMarathonApp(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Manage apps",
	}

	cmd.AddCommand(
		newCmdMarathonAppAdd(ctx),
		newCmdMarathonAppKill(ctx),
		newCmdMarathonAppList(ctx),
		newCmdMarathonAppRemove(ctx),
		newCmdMarathonAppRestart(ctx),
		newCmdMarathonAppShow(ctx),
		newCmdMarathonAppStart(ctx),
		newCmdMarathonAppStop(ctx),
		newCmdMarathonAppUpdate(ctx),
		newCmdMarathonAppVersion(ctx),
	)

	return cmd
}
