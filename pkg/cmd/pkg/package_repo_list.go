package pkg

import (
	"encoding/json"
	"errors"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/cosmos"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

const description = `Print the package repository sources

Possible sources include a local file, HTTPS, and Git.`

const noReposError = "There are currently no repos configured. Please use `dcos package repo add` to add a repo"

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

			// TODO: this follows the python behavior but is this really an error?
			if len(repos.Repositories) == 0 {
				return errors.New(noReposError)
			}

			tableHeader := []string{"NAME", "URI"}
			table := cli.NewTable(ctx.Out(), tableHeader)

			for _, r := range repos.Repositories {
				fields := []string{r.Name, r.Uri}
				table.Append(fields)
			}
			table.Render()

			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	return cmd
}
