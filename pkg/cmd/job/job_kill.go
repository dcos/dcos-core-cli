package job

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

// newCmdJobKill kills a run of a job.
func newCmdJobKill(ctx api.Context) *cobra.Command {
	var all bool
	cmd := &cobra.Command{
		Use:   "kill <job-id> [<run-id>]",
		Short: "Kill a job",
		Args:  cobra.RangeArgs(1, 2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 2 && all == true {
				return fmt.Errorf("cannot accept both a run-id and the --all option")
			}
			if len(args) == 1 && all == false {
				return fmt.Errorf("run-id must be specified or --all option must be set")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := metronomeClient(ctx)
			if err != nil {
				return err
			}

			if len(args) == 2 {
				return client.Kill(args[0], args[1])
			}

			runs, err := client.Runs(args[0])
			if err != nil {
				return err
			}
			for _, run := range runs {
				err = client.Kill(args[0], run.ID)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "Kill all the active runs of this job")
	return cmd
}
