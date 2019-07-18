package node

import (
	"fmt"
	"time"

	"github.com/dcos/dcos-cli/api"
	"github.com/mesos/mesos-go/api/v1/lib"
	"github.com/spf13/cobra"
)

func newCmdNodeDrain(ctx api.Context) *cobra.Command {
	var decommission bool
	var maxGracePeriod time.Duration
	var wait bool
	cmd := &cobra.Command{
		Use:   "drain <mesos-id>",
		Short: "Drain a node so that its tasks get rescheduled",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := mesosClient(ctx)
			if err != nil {
				return err
			}
			err = c.DrainAgent(args[0], maxGracePeriod, decommission)
			if err != nil {
				return err
			}

			if wait {
				fmt.Fprintln(ctx.Out(), "Waiting for the agent to be drained...")
				for range time.Tick(5 * time.Second) {
					agents, err := c.Agents()
					if err != nil {
						return err
					}
					for _, agent := range agents {
						if args[0] == agent.AgentInfo.GetID().Value {
							if agent.GetDrainInfo().GetState() == mesos.DrainState_DRAINED {
								return nil
							}
						}
					}
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&decommission, "decommission", false, "Decommission the agent after having drained it")
	cmd.Flags().DurationVar(&maxGracePeriod, "max-grace-period", 0, "Maximum duration before Mesos will forcefully terminate the agent's tasks")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait until the draining is done")
	return cmd
}
