package task

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/mesos"
	lib "github.com/mesos/mesos-go/api/v1/lib"
	"github.com/spf13/cobra"
)

func newCmdTaskList(ctx api.Context) *cobra.Command {
	var all, jsonOutput, completed, quietOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Print the Mesos tasks in the cluster",
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if all && completed {
				return fmt.Errorf("cannot accept both options --all and --completed")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mesos.NewClientWithContext(ctx)
			if err != nil {
				return err
			}

			tasks, err := client.Tasks()
			if err != nil {
				return err
			}

			if !all {
				tasks = filter(tasks, completed)
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(tasks)
			}

			if quietOutput {
				for _, t := range tasks {
					fmt.Fprintln(ctx.Out(), t.Name)
				}
				return nil
			}

			tableHeader := []string{"NAME", "HOST", "USER", "STATE", "ID", "AGENT ID", "REGION", "ZONE"}
			table := cli.NewTable(ctx.Out(), tableHeader)

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
					if a.AgentInfo.ID.GetValue() == t.AgentID.Value {
						host = a.AgentInfo.Hostname
						region = a.AgentInfo.Domain.FaultDomain.GetRegion().Name
						zone = a.AgentInfo.Domain.FaultDomain.GetZone().Name
					}
				}

				var user string
				for _, f := range frameworks {
					if f.FrameworkInfo.ID.GetValue() == t.FrameworkID.Value {
						user = f.FrameworkInfo.User
					}
				}

				item := []string{
					t.Name,
					host,
					user,
					strings.TrimPrefix(t.State.String(), "TASK_"),
					t.TaskID.Value,
					t.AgentID.Value,
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

func filter(tasks []lib.Task, completed bool) []lib.Task {
	var completedTaskStates = map[string]bool{
		lib.TaskState_name[int32(lib.TASK_FINISHED)]:         true,
		lib.TaskState_name[int32(lib.TASK_KILLED)]:           true,
		lib.TaskState_name[int32(lib.TASK_FAILED)]:           true,
		lib.TaskState_name[int32(lib.TASK_LOST)]:             true,
		lib.TaskState_name[int32(lib.TASK_ERROR)]:            true,
		lib.TaskState_name[int32(lib.TASK_GONE)]:             true,
		lib.TaskState_name[int32(lib.TASK_GONE_BY_OPERATOR)]: true,
		lib.TaskState_name[int32(lib.TASK_DROPPED)]:          true,
		lib.TaskState_name[int32(lib.TASK_UNREACHABLE)]:      true,
		lib.TaskState_name[int32(lib.TASK_UNKNOWN)]:          true,
	}
	var result []lib.Task
	for _, t := range tasks {
		if completed == completedTaskStates[t.State.String()] {
			result = append(result, t)
		}
	}
	return result
}
