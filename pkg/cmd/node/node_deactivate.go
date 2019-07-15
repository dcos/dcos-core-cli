package node

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdNodeDeactivate(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deactivate <mesos-id>",
		Short: "Deactivate a node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet")
		},
	}
	return cmd
}
