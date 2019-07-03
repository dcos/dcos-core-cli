package task

import (
	"encoding/json"
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/mesos"
	"github.com/spf13/cobra"
)

func newCmdTaskList(ctx api.Context) *cobra.Command {
	var all, jsonOutput, completed, quietOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Print the Mesos tasks in the cluster",
		Args:  cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if all && completed {
				return fmt.Errorf("cannot accept both options --all and --completed")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			filters := taskFilters{
				Active:    !completed,
				Completed: all || completed,
			}
			if len(args) == 1 {
				filters.ID = args[0]
			}

			tasks, err := findTasks(ctx, filters)
			if err != nil {
				if jsonOutput {
					// On JSON ouput, we print an empty array instead of erroring-out.
					// This is mainly done for backwards compatibility with the Python CLI.
					fmt.Println("[]")
					return nil
				}
				return err
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(tasks)
			}

			if quietOutput {
				for _, t := range tasks {
					fmt.Fprintln(ctx.Out(), t.ID)
				}
				return nil
			}

			tableHeader := []string{"NAME", "HOST", "USER", "STATE", "ID", "AGENT ID", "REGION", "ZONE"}
			table := cli.NewTable(ctx.Out(), tableHeader)

			client, err := mesos.NewClientWithContext(ctx)
			if err != nil {
				return err
			}

			agents, err := client.Agents()
			if err != nil {
				return err
			}

			frameworks, err := client.Frameworks()
			if err != nil {
				return err
			}

			for _, t := range tasks {
				var host, region, zone string
				for _, a := range agents {
					if a.AgentInfo.ID.GetValue() == t.SlaveID {
						host = a.AgentInfo.Hostname
						region = a.AgentInfo.Domain.FaultDomain.GetRegion().Name
						zone = a.AgentInfo.Domain.FaultDomain.GetZone().Name
					}
				}

				var user string
				for _, f := range frameworks {
					if f.FrameworkInfo.ID.GetValue() == t.FrameworkID {
						user = f.FrameworkInfo.User
					}
				}

				item := []string{
					t.Name,
					host,
					user,
					t.State,
					t.ID,
					t.SlaveID,
					region,
					zone,
				}
				table.Append(item)
			}

			table.Render()
			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "Print completed and in-progress tasks")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	cmd.Flags().BoolVar(&completed, "completed", false, "Print completed tasks")
	cmd.Flags().BoolVarP(&quietOutput, "quiet", "q", false, "Print only IDs of listed services")
	return cmd
}
