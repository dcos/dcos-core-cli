package job

import (
	"encoding/json"
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/metronome"
	"github.com/spf13/cobra"
)

// newCmdJobList lists the jobs.
func newCmdJobList(ctx api.Context) *cobra.Command {
	var jsonOutput bool
	var quietOutput bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all job definitions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := metronomeClient(ctx)
			if err != nil {
				return err
			}

			jobs, err := client.Jobs(
				metronome.EmbedActiveRun(),
				metronome.EmbedSchedule(),
				metronome.EmbedHistorySummary(),
			)
			if err != nil {
				return err
			}

			if quietOutput {
				for _, job := range jobs {
					fmt.Fprintln(ctx.Out(), job.ID)
				}
				return nil
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(jobs)
			}

			table := cli.NewTable(ctx.Out(), []string{"ID", "STATUS", "LAST RUN"})
			for _, job := range jobs {
				lastRunStatus, err := job.LastRunStatus()
				if err != nil {
					return err
				}
				table.Append([]string{job.ID, job.Status(), lastRunStatus})
			}
			table.Render()

			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "returns jobs in json format")
	cmd.Flags().BoolVar(&quietOutput, "quiet", false, "returns only IDs of listed jobs")
	return cmd
}
