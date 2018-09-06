package job

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

// newCmdClusterRun runs a given job right now.
func newCmdJobRun(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "run a job now",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := metronomeClient(ctx)
			if err != nil {
				return err
			}

			_, err = client.RunJob(args[0])
			return err
		},
	}
	return cmd
}
