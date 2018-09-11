package job

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

// newCmdClusterUpdate updates a job.
func newCmdJobUpdate(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <file>",
		Short: "update a job",
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

			if len(job.Schedules) != 0 {
				job.Schedules = nil
			}

			_, err = client.UpdateJob(job)
			return err
		},
	}
	return cmd
}
