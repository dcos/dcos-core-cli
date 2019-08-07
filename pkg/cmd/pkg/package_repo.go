package pkg

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdPackageRepo(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage packages' repositories",
	}

	cmd.AddCommand(
		newCmdPackageRepoAdd(ctx),
		newCmdPackageRepoImport(ctx),
		newCmdPackageRepoList(ctx),
		newCmdPackageRepoRemove(ctx),
	)
	return cmd
}
