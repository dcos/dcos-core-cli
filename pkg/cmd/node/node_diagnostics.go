package node

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

// newCmdNodeMetrics creates the `core node metrics` subcommand.
func newCmdNodeDiagnostics(ctx api.Context) *cobra.Command {
	var cancel bool
	var list bool
	var status bool
	cmd := &cobra.Command{
		Use:   "diagnostics",
		Short: "Use diagnostics bundles",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := ctx.Deprecated("'dcos node diagnostics' command is deprecated since 2.0, but still functional in this release." +
				" Please use the new 'dcos diagnostics' subcommands for bundle creation/download in the future.")
			if err != nil {
				return err
			}
			var subCommand *cobra.Command
			if cancel {
				subCommand = newCmdNodeDiagnosticsCancel(ctx)
			} else if list {
				subCommand = newCmdNodeDiagnosticsList(ctx)
			} else if status {
				subCommand = newCmdNodeDiagnosticsStatus(ctx)
			} else {
				return cmd.Help()
			}

			subCommand.SilenceErrors = true
			subCommand.SilenceUsage = true
			subCommand.SetArgs(ctx.Args()[4:])
			return subCommand.Execute()

		},
	}
	cmd.Flags().BoolVar(&cancel, "cancel", false, "Cancel a running diagnostics job")
	cmd.Flags().MarkDeprecated("cancel", "use the 'cancel' subcommand instead")
	cmd.Flags().Bool("json", false, "Print in json format")
	cmd.Flags().BoolVar(&list, "list", false, "List available diagnostics bundles")
	cmd.Flags().MarkDeprecated("list", "use the 'list' subcommand instead")
	cmd.Flags().BoolVar(&status, "status", false, "Print diagnostics job status")
	cmd.Flags().MarkDeprecated("status", "use the 'status' subcommand instead")

	cmd.AddCommand(
		newCmdNodeDiagnosticsCancel(ctx),
		newCmdNodeDiagnosticsCreate(ctx),
		newCmdNodeDiagnosticsDelete(ctx),
		newCmdNodeDiagnosticsDownload(),
		newCmdNodeDiagnosticsList(ctx),
		newCmdNodeDiagnosticsStatus(ctx),
	)
	return cmd
}
