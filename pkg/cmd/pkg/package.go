package pkg

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

// NewCommand creates the `dcos package` subcommand.
func NewCommand(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "package",
		Short: "Install and manage DC/OS software packages",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			fmt.Fprintln(ctx.ErrOut(), cmd.UsageString())
			return fmt.Errorf("unknown command %s", args[0])
		},
	}

	cmd.AddCommand(
		newCmdPackageDescribe(ctx),
		newCmdPackageInstall(ctx),
		newCmdPackageList(ctx),
		newCmdPackageRepo(ctx),
		newCmdPackageSearch(ctx),
		newCmdPackageUninstall(ctx),
	)

	return cmd
}
