package task

import (
	"os"

	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdTaskAttach(ctx api.Context) *cobra.Command {
	var noStdin bool

	cmd := &cobra.Command{
		Use:   "attach <task>",
		Short: "Attach the CLI to the stdio of an already running task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskIO, err := newTaskIO(ctx, args[0], !noStdin, true, "")
			if err != nil {
				return err
			}

			if err != nil {
				return err
			}

			exitCode, err := taskIO.Attach()
			if err != nil {
				return err
			}
			os.Exit(exitCode)
			return nil
		},
	}

	cmd.Flags().BoolVar(&noStdin, "no-stdin", false, "Don't attach the stdin of the CLI to the task")

	return cmd
}
