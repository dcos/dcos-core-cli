package pkg

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

const description = `Print the package repository sources

Possible sources include a local file, HTTPS, and Git.`

func newCmdPackageRepoList(ctx api.Context) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Print the package repository sources",
		Long:  description,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return invokePythonCLI(ctx)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	return cmd
}
