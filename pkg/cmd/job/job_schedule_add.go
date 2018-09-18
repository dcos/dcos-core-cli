package job

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

// newCmdJobScheduleAdd adds a schedule to a job.
func newCmdJobScheduleAdd(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <job-id> <schedule-file>",
		Short: "add a schedule to a job",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := metronomeClient(ctx)
			if err != nil {
				return err
			}

			reader, err := inputReader(ctx, args[1:])
			if err != nil {
				return err
			}

			schedule, err := parseJSONSchedule(reader)
			if err != nil {
				return err
			}

			_, err = client.AddSchedule(args[0], schedule)
			return err
		},
	}
	return cmd
}
