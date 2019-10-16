package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/dcos/client-go/dcos"
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/cosmos"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

func newCmdPackageList(ctx api.Context) *cobra.Command {
	var cliOnly, jsonOutput bool
	var appID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Print a list of the installed DC/OS packages",
		Args:  cobra.RangeArgs(0, 1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if appID != "" {
				if len(args) > 0 {
					return errors.New("cannot use the flags `--app-id` and an argument at the same time")
				}

				if cliOnly && !jsonOutput {
					return errors.New("cannot use the flags `--app-id` and `--cli` at the same time")
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var list []cosmos.Package
			pkgNames := make(map[string]bool)

			// Get the commands installed locally, those do not have app IDs thus we can skip if one was provided using `app-id`.
			cluster, err := ctx.Cluster()
			if err != nil {
				return err
			}

			plugins := ctx.PluginManager(cluster).Plugins()
			for _, plugin := range plugins {
				packagePath := path.Join(filepath.Dir(plugin.Dir()), "package.json")
				if _, err := os.Stat(packagePath); err == nil {
					file, err := ioutil.ReadFile(packagePath)
					if err != nil {
						return err
					}

					pkg := cosmos.Package{}
					err = json.Unmarshal(file, &pkg)
					if err != nil {
						return err
					}

					if pkg.Name != "" {
						if len(pkg.Apps) == 0 && !strings.Contains(pkg.Name, "/") {
							pkg.Apps = append(pkg.Apps, "/"+pkg.Name)
						}

						pkg.Command = &dcos.CosmosPackageCommand{
							Name: pkg.Name,
						}

						if !pkgNames[pkg.Name] {
							pkgNames[pkg.Name] = true
							list = append(list, pkg)
						}
					}
				}
			}

			// Get the list of packages from Cosmos.
			if !cliOnly {
				c, err := cosmos.NewClient(ctx, pluginutil.HTTPClient(""))
				if err != nil {
					return err
				}

				cosmosPackages, err := c.PackageList()
				if err != nil {
					return err
				}

				for _, cosmosPkg := range cosmosPackages {
					if !pkgNames[cosmosPkg.Name] {
						pkgNames[cosmosPkg.Name] = true
						list = append(list, cosmosPkg)
					}
				}
			}

			// Filter the packages to only keep the ones with a matching app name.
			var filteredList []cosmos.Package
			if appID != "" || len(args) > 0 {
				for _, pkg := range list {
					for _, app := range pkg.Apps {
						if (len(args) > 0 && strings.Contains(app, appID)) || (app == appID) {
							filteredList = append(filteredList, pkg)
							break
						}
					}
				}
				if !jsonOutput && (len(filteredList) == 0) {
					return errors.New("cannot find packages matching the provided filter")
				}
				list = filteredList
			}

			if jsonOutput {
				if len(list) == 0 {
					fmt.Fprintln(ctx.Out(), "[]")
					return nil
				}

				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(&list)
			}

			if len(list) == 0 {
				return errors.New("there are currently no installed packages")
			}

			table := cli.NewTable(ctx.Out(), []string{"NAME", "VERSION", "SELECTED", "APP", "COMMAND", "DESCRIPTION"})
			for _, p := range list {
				apps := "---"
				if len(p.Apps) > 0 {
					apps = strings.Join(p.Apps, "\n")
				}
				command := "---"
				if p.Command != nil && p.Command.Name != "" {
					command = p.Command.Name
				}

				description := p.Description
				if strings.Index(description, "\n") > -1 {
					i := strings.Index(description, "\n")
					description = description[:i] + "..."
				}
				if utf8.RuneCountInString(description) > 70 {
					description = description[0:66] + "..."
				}

				table.Append([]string{p.Name, p.Version, strconv.FormatBool(p.Selected), apps, command, description})
			}
			table.Render()

			return nil
		},
	}
	cmd.Flags().StringVar(&appID, "app-id", "", "The application ID")
	cmd.Flags().BoolVar(&cliOnly, "cli", false, "Command line only")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	return cmd
}
