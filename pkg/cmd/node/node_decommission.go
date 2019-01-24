package node

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdNodeDecommission(ctx api.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "decommission",
		Short: "Mark an agent as gone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return mesosClient().MarkAgentGone(args[0])
		},
	}
}
