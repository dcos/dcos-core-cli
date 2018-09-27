package job

import (
	"encoding/json"

	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

// newCmdJobShow displays a job definition.
func newCmdJobShow(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <job-id>",
		Short: "Displays a job definition",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := metronomeClient(ctx)
			if err != nil {
				return err
			}

			job, err := client.Job(args[0])
			if err != nil {
				return err
			}

			enc := json.NewEncoder(ctx.Out())
			enc.SetIndent("", "    ")
			return enc.Encode(job)
		},
	}

	cmd.AddCommand(
		newCmdJobShowRuns(ctx),
	)

	return cmd
}
