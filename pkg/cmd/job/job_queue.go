package job

import (
	"encoding/json"
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/metronome"
	"github.com/spf13/cobra"
)

// newCmdJobQueue displays all queued job runs.
func newCmdJobQueue(ctx api.Context) *cobra.Command {
	var jsonOutput bool
	var quietOutput bool
	cmd := &cobra.Command{
		Use:   "queue",
		Short: "Show job runs that are queued",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := metronomeClient(ctx)
			if err != nil {
				return err
			}

			var queued []metronome.Queue
			if len(args) == 1 {
				queued, err = client.Queued(args[0])
			} else {
				queued, err = client.Queued("")
			}
			if err != nil {
				return err
			}

			if quietOutput {
				for _, queue := range queued {
					for _, run := range queue.Runs {
						fmt.Fprintln(ctx.Out(), run.ID)
					}
				}
				return nil
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(queued)
			}

			table := cli.NewTable(ctx.Out(), []string{"JOB ID", "RUN ID"})
			for _, queue := range queued {
				for _, run := range queue.Runs {
					table.Append([]string{queue.JobID, run.ID})
				}
			}
			table.Render()

			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "returns queued job runs in json format")
	cmd.Flags().BoolVar(&quietOutput, "quiet", false, "returns only IDs of queued job runs")
	return cmd
}
