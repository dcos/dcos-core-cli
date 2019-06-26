package task

import (
	"encoding/json"
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/metrics"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

func newCmdTaskMetricsDetails(ctx api.Context) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "details <task-id>",
		Short: "Print a table of all the metrics for a given task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			task, err := findTask(ctx, args[0])
			if err != nil {
				return err
			}
			status := task.Statuses[0]
			containerID := status.ContainerStatus.GetContainerID().Value

			c := metrics.NewClient(pluginutil.HTTPClient(""))
			taskMetrics, err := c.Task(task.AgentID.Value, containerID)
			if err != nil {
				return err
			}
			appMetrics, err := c.App(task.AgentID.Value, containerID)
			if err != nil {
				return err
			}

			if taskMetrics == nil && appMetrics == nil {
				return fmt.Errorf("No metrics found for task '%s'", task.TaskID.Value)
			}
			datapoints := []metrics.Datapoint{}
			if taskMetrics != nil {
				datapoints = append(datapoints, taskMetrics.Datapoints...)
			}

			if appMetrics != nil {
				datapoints = append(datapoints, appMetrics.Datapoints...)
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(datapoints)
			}

			table := cli.NewTable(ctx.Out(), []string{"NAME", "VALUE"})
			for _, datapoint := range datapoints {
				value := ""
				if datapoint.Value == float64(int(datapoint.Value)) {
					value = fmt.Sprintf("%.0f", datapoint.Value)
				} else {
					value = fmt.Sprintf("%.2f", datapoint.Value)
				}
				table.Append([]string{datapoint.Name, value})
			}
			table.Render()
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	return cmd
}
