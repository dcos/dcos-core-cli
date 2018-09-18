package job

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

// newCmdJobScheduleUpdate updates a schedule of a job.
func newCmdJobScheduleUpdate(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <job-id> <schedule-file>",
		Short: "update a schedule of a job",
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

			if schedule.ID == "" {
				return fmt.Errorf("field 'id' needs to be set in schedule JSON")
			}

			_, err = client.UpdateSchedule(args[0], schedule)
			return err
		},
	}
	return cmd
}
