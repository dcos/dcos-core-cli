package job

import (
	"encoding/json"
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/metronome"
	"github.com/spf13/cobra"
)

// newCmdJobShowRuns displays the currents job runs for a given <run-id>.
func newCmdJobShowRuns(ctx api.Context) *cobra.Command {
	var runID string
	var quietOutput bool
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "runs <job-id>",
		Short: "Show the successful and failure runs of a job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := metronomeClient(ctx)
			if err != nil {
				return err
			}

			var runs []metronome.Run
			if runID != "" {
				run, runErr := client.Run(args[0], runID)
				if runErr != nil {
					return runErr
				}
				runs = append(runs, *run)
			} else {
				runs, err = client.Runs(args[0])
				if err != nil {
					return err
				}
			}

			if quietOutput {
				for _, run := range runs {
					fmt.Fprintln(ctx.Out(), run.ID)
				}
				return nil
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(runs)
			}

			table := cli.NewTable(ctx.Out(), []string{"TASK ID", "JOB ID", "STARTED AT"})
			for _, run := range runs {
				table.Append([]string{run.ID, args[0], run.CreatedAt})
			}
			table.Render()

			return nil
		},
	}

	cmd.Flags().StringVar(&runID, "run-id", "", "Show run for a given <run-id>")
	cmd.Flags().BoolVarP(&quietOutput, "quiet", "q", false, "Print only IDs of listed runs")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	return cmd
}
