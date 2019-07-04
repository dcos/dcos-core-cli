package task

import (
	"os"

	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdTaskExec(ctx api.Context) *cobra.Command {
	var interactive, tty bool
	var user string

	cmd := &cobra.Command{
		Use:   "exec [flags] <task> <cmd> [<args>...]",
		Short: "Launch a process inside of a container for a task",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskIO, err := newTaskIO(ctx, args[0], interactive, tty, user)
			if err != nil {
				return err
			}

			exitCode, err := taskIO.Exec(args[1], args[2:]...)
			if err != nil {
				return err
			}
			os.Exit(exitCode)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Attach a STDIN stream to the remote command for an interactive session")
	cmd.Flags().BoolVarP(&tty, "tty", "t", false, "Attach a tty to the remote stream.")
	cmd.Flags().StringVarP(&user, "user", "u", "", "Run as the given user")
	cmd.Flags().SetInterspersed(false)
	cmd.DisableFlagsInUseLine = true
	return cmd
}
