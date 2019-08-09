package pkg

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdPackageRepoAdd(ctx api.Context) *cobra.Command {
	var index int

	cmd := &cobra.Command{
		Use:   "add <repo-name> <repo-url>",
		Short: "Add a package repository to DC/OS",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return invokePythonCLI(ctx)
		},
	}
	cmd.Flags().IntVar(&index, "index", 0, "Numerical position in the package repository list. Package repositories are searched in descending order.")
	return cmd
}
