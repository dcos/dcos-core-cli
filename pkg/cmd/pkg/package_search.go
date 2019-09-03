package pkg

import (
	"encoding/json"
	"errors"
	"strconv"
	"unicode/utf8"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/cosmos"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

func newCmdPackageSearch(ctx api.Context) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search the package repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := cosmos.NewClient(ctx, pluginutil.HTTPClient(""))
			if err != nil {
				return err
			}

			query := ""
			if len(args) == 1 {
				query = args[0]
			}
			searchResult, err := c.PackageSearch(query)
			if err != nil {
				return err
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(searchResult)
			}

			if query != "" && len(searchResult.Packages) == 0 {
				return errors.New("no packages found")
			}

			table := cli.NewTable(ctx.Out(), []string{"NAME", "VERSION", "SELECTED", "FRAMEWORK", "DESCRIPTION"})
			for _, cosmosPackage := range searchResult.Packages {
				description := cosmosPackage.Description
				if utf8.RuneCountInString(cosmosPackage.Description) >= 80 {
					description = description[0:76] + "..."
				}
				table.Append([]string{
					cosmosPackage.Name,
					cosmosPackage.CurrentVersion,
					strconv.FormatBool(cosmosPackage.Selected),
					strconv.FormatBool(cosmosPackage.Framework),
					description,
				})
			}
			table.Render()

			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	return cmd
}
