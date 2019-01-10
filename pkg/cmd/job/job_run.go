package job

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

// newCmdClusterRun runs a given job right now.
func newCmdJobRun(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a job now",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := metronomeClient(ctx)
			if err != nil {
				return err
			}

			run, err := client.RunJob(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(ctx.Out(), run.ID)
			return nil
		},
	}
	return cmd
}
