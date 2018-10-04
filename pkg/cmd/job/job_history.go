package job

import (
	"encoding/json"
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/metronome"
	"github.com/spf13/cobra"
)

// newCmdJobHistory displays the history of the runs of a job.
func newCmdJobHistory(ctx api.Context) *cobra.Command {
	var jsonOutput bool
	var quietOutput bool
	var last bool
	var failures bool
	cmd := &cobra.Command{
		Use:   "history",
		Short: "View your job run history",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := metronomeClient(ctx)
			if err != nil {
				return err
			}

			job, err := client.Job(
				args[0],
				metronome.EmbedActiveRun(),
				metronome.EmbedSchedule(),
				metronome.EmbedHistory(),
			)
			if err != nil {
				return err
			}

			if job.History == nil {
				return fmt.Errorf("no history available for this job")
			}

			runs := job.History.SuccessfulFinishedRuns
			if failures {
				runs = job.History.FailedRuns
			}

			if quietOutput {
				for _, run := range runs {
					fmt.Fprintln(ctx.Out(), run.ID)
					if last {
						return nil
					}
				}
				return nil
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(runs)
			}

			fmt.Fprintln(ctx.Out(), historyMessage(job, failures))
			table := cli.NewTable(ctx.Out(), []string{"ID", "STARTED", "FINISHED"})
			for _, run := range runs {
				table.Append([]string{run.ID, run.CreatedAt, run.FinishedAt})
			}
			table.Render()

			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "returns history in json format")
	cmd.Flags().BoolVar(&quietOutput, "quiet", false, "returns only IDs of listed jobs")
	cmd.Flags().BoolVar(&last, "last", false, "returns history for last run")
	cmd.Flags().BoolVar(&failures, "show-failures", false, "show failed runs of this job")
	return cmd
}

func historyMessage(job *metronome.Job, failures bool) string {
	if failures {
		return fmt.Sprintf(
			`"%s"  Failure runs: %d Last Failure: %s`,
			job.ID, job.History.FailureCount, job.History.LastFailureAt,
		)
	}

	return fmt.Sprintf(
		`"%s"  Successful runs: %d Last Success: %s`,
		job.ID, job.History.SuccessCount, job.History.LastSuccessAt,
	)
}
