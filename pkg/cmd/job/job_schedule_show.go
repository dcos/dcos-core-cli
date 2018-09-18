package job

import (
	"encoding/json"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/spf13/cobra"
)

// newCmdJobScheduleShow lists all the schedules of a job.
func newCmdJobScheduleShow(ctx api.Context) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "show <job-id>",
		Short: "show the schedule of a job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := metronomeClient(ctx)
			if err != nil {
				return err
			}

			schedules, err := client.Schedules(args[0])
			if err != nil {
				return err
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(schedules)
			}

			table := cli.NewTable(ctx.Out(), []string{"ID", "CRON", "ENABLED", "NEXT RUN", "CONCURRENCY POLICY"})
			for _, schedule := range schedules {
				enabled := "False"
				if schedule.Enabled {
					enabled = "True"
				}
				table.Append([]string{schedule.ID, schedule.Cron, enabled, schedule.NextRunAt, schedule.ConcurrencyPolicy})
			}
			table.Render()
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "returns schedules in json format")
	return cmd
}
