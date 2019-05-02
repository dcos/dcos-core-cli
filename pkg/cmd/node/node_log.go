package node

import (
	"fmt"
	"strings"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/logs"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

func newCmdNodeLog(ctx api.Context) *cobra.Command {
	var component, mesosID, output string
	var filters []string
	var follow, leader bool
	var lines int

	cmd := &cobra.Command{
		Use:   "log",
		Short: "Print logs for the leading master node or agent nodes",
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !leader && mesosID == "" {
				return fmt.Errorf("'--leader' or '--mesos-id' must be provided")
			} else if leader && mesosID != "" {
				return fmt.Errorf("unable to use leader and mesos id at the same time")
			}

			for _, filter := range filters {
				if len(strings.Split(filter, ":")) != 2 {
					return fmt.Errorf("invalid filter argument %s, must be --filter=key:value", filter)
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			route := ""
			if leader {
				route = "/leader/mesos"
			} else {
				route = "/agent/" + mesosID
			}

			service := ""
			if component != "" {
				service = fmt.Sprintf("/%s.service", component)
			}

			if mesosID != "" {
				c, err := mesosClient(ctx)
				if err != nil {
					return err
				}
				agents, err := c.Agents()
				if err != nil {
					return err
				}
				for i, agent := range agents {
					if mesosID == agent.AgentInfo.GetID().Value {
						break
					}
					if i == len(agents)-1 {
						return fmt.Errorf("agent '%s' not found", mesosID)
					}
				}

			}

			client := logs.NewClient(pluginutil.HTTPClient(""), ctx.Out())

			opts := logs.Options{
				Filters: filters,
				Follow:  follow,
				Format:  output,
				Skip:    -1 * lines,
			}

			return client.PrintComponent(route, service, opts)
		},
	}
	cmd.Flags().StringVar(&component, "component", "", "Show DC/OS component logs")
	cmd.Flags().StringArrayVar(&filters, "filter", nil, "Filter logs by field and value. Filter must be a string separated by colon. For example: --filter _PID:0 --filter _UID:1")
	cmd.Flags().BoolVar(&follow, "follow", false, "Dynamically update the log")
	cmd.Flags().BoolVar(&leader, "leader", false, "The leading master")
	cmd.Flags().IntVar(&lines, "lines", 10, "Print the N last lines. 10 is the default")
	cmd.Flags().StringVar(&mesosID, "mesos-id", "", "The agent ID of a node")
	cmd.Flags().StringVarP(&output, "output", "o", "short", "Format log message output")
	return cmd
}
