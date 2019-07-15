package node

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdNodeDrain(ctx api.Context) *cobra.Command {
	var decommission bool
	var timeout int
	var wait bool
	cmd := &cobra.Command{
		Use:   "drain <mesos-id>",
		Short: "Drain a node so that its tasks get rescheduled",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet")
		},
	}
	cmd.Flags().BoolVar(&decommission, "decommission", false, "Decommission the agent after having drained it")
	cmd.Flags().IntVar(&timeout, "timeout", 0, "Timeout to do the request")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait until the draining is done")
	return cmd
}
