package job

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

// newCmdJobRemove removes a given job.
func newCmdJobRemove(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "remove a job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := metronomeClient(ctx)
			if err != nil {
				return err
			}

			return client.RemoveJob(args[0])
		},
	}
	return cmd
}
