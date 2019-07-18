package node

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdNodeReactivate(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reactivate <mesos-id>",
		Short: "Reactivate a drained/deactivated node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := mesosClient(ctx)
			if err != nil {
				return err
			}
			return c.ReactivateAgent(args[0])
		},
	}
	return cmd
}
