package pkg

import (
	"encoding/json"
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cosmos"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
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
			c, err := cosmos.NewClient(ctx, pluginutil.HTTPClient(""))
			if err != nil {
				return err
			}

			repos, err := c.PackageListRepo()
			if err != nil {
				return err
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(repos)
			}

			for _, r := range repos.Repositories {
				fmt.Fprintf(ctx.Out(), "%s: %s\n", r.Name, r.Uri)
			}

			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	return cmd
}
