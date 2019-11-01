package app

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/python"
	"github.com/spf13/cobra"
)

func newCmdMarathonAppVersion(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Manage Marathon app versions.",
	}

	cmd.AddCommand(
		newCmdMarathonAppVersionList(ctx),
	)

	return cmd
}

func newCmdMarathonAppVersionList(ctx api.Context) *cobra.Command {
	var maxCount int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the version history of an application.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return python.InvokePythonCLI(ctx)
		},
	}

	cmd.Flags().IntVar(&maxCount, "max-count", 0, "Maximum number of entries to fetch and return.")

	return cmd
}
