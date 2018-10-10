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
		Short: "Show all job definitions",
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
				table.Append([]string{job.ID, job.Status(), job.LastRunStatus()})
			}
			table.Render()

			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	cmd.Flags().BoolVarP(&quietOutput, "quiet", "q", false, "Print only IDs of listed jobs")
	return cmd
}
