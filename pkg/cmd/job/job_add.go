package job

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/metronome"
	"github.com/spf13/cobra"
)

// newCmdClusterAdd creates a new job.
func newCmdJobAdd(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <file>",
		Short: "add a job",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := metronomeClient(ctx)
			if err != nil {
				return err
			}

			// Handling input from file or stdin
			reader, err := inputReader(ctx, args)
			if err != nil {
				return err
			}

			job, err := parseJSONJob(reader)
			if err != nil {
				return err
			}

			if job.ID == "" {
				return fmt.Errorf("jobs JSON requires an ID")
			}

			// Checking for schedule to upload it separately
			schedules := make([]metronome.Schedule, len(job.Schedules))
			if len(job.Schedules) != 0 {
				copy(schedules, job.Schedules)
				job.Schedules = nil
			}

			_, err = client.AddJob(job)
			if err != nil {
				return err
			}

			if len(schedules) != 0 {
				_, err = client.AddSchedule(job.ID, &schedules[0])
			}
			return err
		},
	}
	return cmd
}
