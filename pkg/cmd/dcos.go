package cmd

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cmd/job"
	"github.com/spf13/cobra"
)

// NewDCOSCommand creates the `dcos` command with all the available subcommands.
func NewDCOSCommand(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use: "dcos",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmd.SilenceUsage = true
		},
	}

	cmd.AddCommand(
		job.NewCommand(ctx),
	)

	return cmd
}
