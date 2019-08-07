package pkg

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdPackageRepoImport(ctx api.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "import <repo-file>",
		Short: "Import a file containing a package repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return invokePythonCLI(ctx)
		},
	}
}
