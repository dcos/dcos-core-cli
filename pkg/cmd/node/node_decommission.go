package node

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdNodeDecommission(ctx api.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "decommission <mesos-id>",
		Short: "Mark an agent as gone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := mesosClient(ctx)
			if err != nil {
				return err
			}
			return c.MarkAgentGone(args[0])
		},
	}
}
