package group

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func NewCmdMarathonGroup(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Manage groups.",
	}

	cmd.AddCommand(
		newCmdMarathonGroupAdd(ctx),
		newCmdMarathonGroupList(ctx),
		newCmdMarathonGroupRemove(ctx),
		newCmdMarathonGroupScale(ctx),
		newCmdMarathonGroupShow(ctx),
		newCmdMarathonGroupUpdate(ctx),
	)

	return cmd
}
