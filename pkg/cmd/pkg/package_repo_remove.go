package pkg

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cosmos"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

func newCmdPackageRepoRemove(ctx api.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <repo-names>...",
		Short: "Remove a package repository from DC/OS",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := cosmos.NewClient(ctx, pluginutil.HTTPClient(""))
			if err != nil {
				return err
			}

			for _, arg := range args {
				err = c.PackageDeleteRepo(arg)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
}
