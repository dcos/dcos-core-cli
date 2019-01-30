package node

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

func newCmdNodeMetricsDetails(ctx api.Context) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "details <mesos-id>",
		Short: "Print details of the metrics of an agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			node, err := metrics.NewClient(pluginutil.HTTPClient("")).Node(args[1])
			if err != nil {
				return err
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(node.Datapoints)
			}

			table := cli.NewTable(ctx.Out(), []string{"NAME", "VALUE", "TAGS"})
			for _, datapoint := range node.Datapoints {
				value := ""
				switch datapoint.Unit {
				case "bytes":
					value = fmt.Sprintf("%.2fGiB", datapoint.Value/math.Pow(10, 9))
				case "percent":
					value = fmt.Sprintf("%.2f%%", datapoint.Value)
				default:
					value = fmt.Sprintf("%.0f", datapoint.Value)
				}

				tags := ""
				for key, val := range datapoint.Tags {
					if tags != "" {
						tags += ", "
					}
					tags += fmt.Sprintf("%s: %s", key, val)
				}

				table.Append([]string{datapoint.Name, value, tags})
			}
			table.Render()
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	return cmd
}
