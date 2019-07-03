package task

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/metrics"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

func newCmdTaskMetricsSummary(ctx api.Context) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "summary <task-id>",
		Short: "Print a table of the key metrics for a given task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filters := taskFilters{
				Active: true,
				ID:     args[0],
			}
			task, err := findTask(ctx, filters)
			if err != nil {
				return err
			}
			status := task.Statuses[0]
			containerID := status.ContainerStatus.ContainerID.Value

			c := metrics.NewClient(pluginutil.HTTPClient(""))
			taskMetrics, err := c.Task(task.SlaveID, containerID)
			if err != nil {
				return err
			}

			if taskMetrics == nil {
				return fmt.Errorf("No metrics found for task '%s'", task.ID)
			}

			filteredDatapoints := []metrics.Datapoint{}
			summaryDatapoints := map[string]float64{
				"cpus.user_time_secs":      0,
				"cpus.system_time_secs":    0,
				"cpus.throttled_time_secs": 0,
				"mem.limit_bytes":          0,
				"mem.total_bytes":          0,
				"disk.used_bytes":          0,
				"disk.limit_bytes":         0,
			}

			// Filter the datapoints.
			for _, datapoint := range taskMetrics.Datapoints {
				if _, ok := summaryDatapoints[datapoint.Name]; ok {
					filteredDatapoints = append(filteredDatapoints, datapoint)
					summaryDatapoints[datapoint.Name] = datapoint.Value
				}
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(filteredDatapoints)
			}

			table := cli.NewTable(ctx.Out(), []string{"CPU", "MEM", "DISK"})

			cpuUsed := summaryDatapoints["cpus.user_time_secs"] + summaryDatapoints["cpus.system_time_secs"]
			cpuTotal := cpuUsed + summaryDatapoints["cpus.throttled_time_secs"]
			cpuPercent := "N/A"
			if cpuTotal != 0 {
				cpuPercent = fmt.Sprintf("%.2f%%", (cpuUsed/cpuTotal)*100)
			}

			memUsed := summaryDatapoints["mem.file_bytes"]
			memTotal := summaryDatapoints["mem.total_bytes"]
			memPercent := "N/A"
			if memTotal != 0 {
				memPercent = fmt.Sprintf("%.2f%%", (memUsed/memTotal)*100)
			}

			diskUsed := summaryDatapoints["disk.used_bytes"]
			diskLimit := summaryDatapoints["disk.limit_bytes"]
			diskPercent := "N/A"
			if diskLimit != 0 {
				diskPercent = fmt.Sprintf("%.2f%%", (diskUsed/diskLimit)*100)
			}

			table.Append([]string{
				fmt.Sprintf("%.2f (%s)", cpuUsed, cpuPercent),
				fmt.Sprintf("%.2fGiB (%s)", memUsed/math.Pow(10, 9), memPercent),
				fmt.Sprintf("%.2fGiB (%s)", diskUsed, diskPercent),
			})
			table.Render()
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	return cmd
}
