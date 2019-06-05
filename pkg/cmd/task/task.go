package task

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/spf13/cobra"
)

// NewCommand creates the `core service` subcommand.
func NewCommand(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task <task-id>",
		Short: "Manage DC/OS tasks",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, ok := ctx.EnvLookup(cli.EnvStrictDeprecations)
			if !ok {
				ctx.Deprecated("Getting the list of tasks from `dcos task` is deprecated. Please use `dcos task list`.")
				listCmd := newCmdTaskList(ctx)
				// Execute by default would use os.Args[1:], which is everything after `dcos ...`.
				// We need all command line arguments after `dcos service ...`.
				listCmd.SetArgs(ctx.Args()[2:])
				listCmd.SilenceErrors = true
				listCmd.SilenceUsage = true
				return listCmd.Execute()
			}
			return cmd.Help()
		},
	}
	cmd.Flags().Bool("all", false, "Print completed and in-progress tasks")
	cmd.Flags().Bool("json", false, "Print in json format")
	cmd.Flags().Bool("completed", false, "Print completed tasks")

	cmd.AddCommand(
		newCmdTaskAttach(ctx),
		newCmdTaskList(ctx),
		newCmdTaskLog(ctx),
		newCmdTaskLs(ctx),
	)
	return cmd
}
