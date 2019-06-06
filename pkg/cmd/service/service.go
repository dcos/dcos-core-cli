package service

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/spf13/cobra"
)

// NewCommand creates the `core service` subcommand.
func NewCommand(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Manage DC/OS services",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				_, ok := ctx.EnvLookup(cli.EnvStrictDeprecations)
				if !ok {
					ctx.Deprecated("Getting the list of services from `dcos service` is deprecated. Please use `dcos service list`.")
					listCmd := newCmdServiceList(ctx)
					// Execute by default would use os.Args[1:], which is everything after `dcos ...`.
					// We need all command line arguments after `dcos service ...`.
					listCmd.SetArgs(ctx.Args()[2:])
					listCmd.SilenceErrors = true
					listCmd.SilenceUsage = true
					return listCmd.Execute()
				}
				return cmd.Help()
			}
			fmt.Fprintln(ctx.ErrOut(), cmd.UsageString())
			return fmt.Errorf("unknown command %s", args[0])
		},
	}
	cmd.Flags().Bool("json", false, "Print in json format")
	cmd.Flags().Bool("completed", false, "Print completed and active services")
	cmd.Flags().Bool("inactive", false, "Print inactive and active services")

	cmd.AddCommand(
		newCmdServiceList(ctx),
		newCmdServiceLog(ctx),
		newCmdServiceShutdown(ctx),
	)
	return cmd
}
