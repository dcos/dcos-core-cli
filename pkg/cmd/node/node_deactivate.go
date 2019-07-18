package node

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdNodeDeactivate(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deactivate <mesos-id>",
		Short: "Deactivate a node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := mesosClient(ctx)
			if err != nil {
				return err
			}
			return c.DeactivateAgent(args[0])
		},
	}
	return cmd
}
