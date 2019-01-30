package node

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/metrics"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

func newCmdNodeMetricsSummary(ctx api.Context) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "summary <mesos-id>",
		Short: "Print summary of the metrics of an agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			node, err := metrics.NewClient(pluginutil.HTTPClient("")).Node(args[1])
			if err != nil {
				return err
			}

			filteredDatapoints := []metrics.Datapoint{}
			summaryDatapoints := map[string]float64{
				"load.1min":                 0,
				"cpu.total":                 0,
				"memory.total":              0,
				"memory.free":               0,
				"filesystem.capacity.total": 0,
				"filesystem.capacity.used":  0,
			}

			// Filter the datapoints.
			for _, datapoint := range node.Datapoints {
				if _, ok := summaryDatapoints[datapoint.Name]; ok {
					// Special case for filesystem's datapoints as there are multiple with different paths.
					if strings.Contains(datapoint.Name, "filesystem.capacity.") && datapoint.Tags != nil {
						if val, ok := datapoint.Tags["path"]; ok && val != "/" {
							continue
						}
					}
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
			memUsed := summaryDatapoints["memory.total"] - summaryDatapoints["memory.free"]
			percentMemUsed := memUsed / summaryDatapoints["memory.total"] * 100
			percentDiskUsed := summaryDatapoints["filesystem.capacity.used"] / summaryDatapoints["filesystem.capacity.total"] * 100
			table.Append([]string{
				fmt.Sprintf("%.2f (%.2f%%)", summaryDatapoints["load.1min"], summaryDatapoints["cpu.total"]),
				fmt.Sprintf("%.2fGiB (%.2f%%)", memUsed/math.Pow(10, 9), percentMemUsed),
				fmt.Sprintf("%.2fGiB (%.2f%%)", summaryDatapoints["filesystem.capacity.used"]/math.Pow(10, 9), percentDiskUsed),
			})
			table.Render()
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	return cmd
}
