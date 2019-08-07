package pkg

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdPackageRepoRemove(ctx api.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <repo-names>...",
		Short: "Remove a package repository from DC/OS",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return invokePythonCLI(ctx)
		},
	}
}
