package service

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/mesos"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func newCmdServiceList(ctx api.Context) *cobra.Command {
	var jsonOutput, completed, inactive bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Show DC/OS services",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mesos.NewClientWithContext(ctx)
			if err != nil {
				return err
			}

			state, err := client.State()
			if err != nil {
				return err
			}

			frameworks := state.Frameworks
			if !inactive {
				frameworks = filterInactiveFrameworks(frameworks)
			}

			if completed {
				frameworks = append(frameworks, state.CompletedFrameworks...)
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(frameworks)
			}

			tableHeader := []string{"NAME", "HOST", "ACTIVE", "TASKS", "CPU", "MEM", "DISK", "ID"}
			table := cli.NewTable(ctx.Out(), tableHeader)

			for _, f := range frameworks {
				item := []string{
					f.Name,
					f.Hostname,
					tablewriter.ConditionString(f.Active, "True", "False"),
					strconv.Itoa(len(f.Tasks)),
					fmt.Sprintf("%.1f", f.Resources.CPUs),
					fmt.Sprintf("%.1f", f.Resources.Mem),
					fmt.Sprintf("%.1f", f.Resources.Disk),
					f.ID,
				}
				table.Append(item)
			}

			table.Render()
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	cmd.Flags().BoolVar(&completed, "completed", false, "Print completed and active services")
	cmd.Flags().BoolVar(&inactive, "inactive", false, "Print inactive and active services")
	return cmd
}

func filterInactiveFrameworks(frameworks []mesos.Framework) []mesos.Framework {
	var result []mesos.Framework
	for _, f := range frameworks {
		if f.Active {
			result = append(result, f)
		}
	}
	return result
}
