package diagnostics

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

// NewCommand creates and returns a diagnostics command with its subcommands
// already added.
func NewCommand(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diagnostics",
		Short: "Create and manage DC/OS diagnostics bundles",
	}
	cmd.AddCommand(
		newDiagnosticsListCommand(ctx),
		newDiagnosticsDownloadCommand(ctx),
		newDiagnosticsCreateCommand(ctx),
		newDiagnosticsDeleteCommand(ctx),
		newDiagnosticsWaitCommand(ctx),
	)
	return cmd
}
