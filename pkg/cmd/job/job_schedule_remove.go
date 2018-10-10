package job

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

// newCmdJobScheduleRemove removes a schedule from a job.
func newCmdJobScheduleRemove(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <job-id> <schedule-id>",
		Short: "Remove a schedule of a job",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := metronomeClient(ctx)
			if err != nil {
				return err
			}

			return client.RemoveSchedule(args[0], args[1])
		},
	}
	return cmd
}
