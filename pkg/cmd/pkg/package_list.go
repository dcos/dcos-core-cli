package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
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

type listOptions struct {
	appID string
	query string

	cliOnly    bool
	jsonOutput bool
}

func newCmdPackageList(ctx api.Context) *cobra.Command {
	var opts listOptions

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Print a list of the installed DC/OS packages",
		Args:  cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.appID != "" {
				if len(args) > 0 {
					return errors.New("cannot use the flags `--app-id` and an argument at the same time")
				}

				if opts.cliOnly && !opts.jsonOutput {
					return errors.New("cannot use the flags `--app-id` and `--cli` at the same time")
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				opts.query = args[0]
			}
			c, err := cosmos.NewClient(ctx, pluginutil.HTTPClient(""))
			if err != nil {
				return err
			}
			return listPackages(ctx, opts, c)
		},
	}
	cmd.Flags().StringVar(&opts.appID, "app-id", "", "The application ID")
	cmd.Flags().BoolVar(&opts.cliOnly, "cli", false, "Command line only")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Print options json format")
	return cmd
}

func listPackages(ctx api.Context, opts listOptions, c cosmos.Client) error {
	packages, err := getPackagesInstalledLocally(ctx)
	if err != nil {
		return fmt.Errorf("cannot read cluster data: %s", err)
	}

	if !opts.cliOnly {
		cosmosPackages, err := c.PackageList()
		if err != nil {
			return err
		}

		for _, pkg := range cosmosPackages {
			id := identifier(pkg)
			p, ok := packages[id]
			if !ok {
				packages[id] = pkg
			} else {
				p.Apps = appendIfMissing(p.Apps, pkg.Apps...)
				packages[id] = p
			}
		}
	}

	filter := func(app string) bool { return true }
	if opts.appID != "" {
		filter = func(app string) bool { return opts.appID == app }
	}
	if opts.query != "" {
		filter = func(app string) bool { return strings.Contains(app, opts.query) }
	}
	list := filterPackages(packages, filter)

	if opts.jsonOutput {
		enc := json.NewEncoder(ctx.Out())
		enc.SetIndent("", "    ")
		return enc.Encode(&list)
	}

	if len(list) == 0 {
		return errors.New("cannot find packages matching the provided filter")
	}

	renderTable(ctx, list)

	return nil
}

func getPackagesInstalledLocally(ctx api.Context) (map[id]cosmos.Package, error) {
	cluster, err := ctx.Cluster()
	if err != nil {
		return nil, err
	}
	packages := make(map[id]cosmos.Package)
	plugins := ctx.PluginManager(cluster).Plugins()
	for _, plugin := range plugins {
		packagePath := path.Join(filepath.Dir(plugin.Dir()), "package.json")
		if _, err := os.Stat(packagePath); err == nil {
			file, err := ioutil.ReadFile(packagePath)
			if err != nil {
				return nil, err
			}

			pkg := cosmos.Package{}
			err = json.Unmarshal(file, &pkg)
			if err != nil {
				return nil, err
			}

			if pkg.Name != "" {
				if len(pkg.Apps) == 0 && !strings.Contains(pkg.Name, "/") {
					pkg.Apps = append(pkg.Apps, "/"+pkg.Name)
				}

				pkg.Command = &dcos.CosmosPackageCommand{
					Name: pkg.Name,
				}

				packages[identifier(pkg)] = pkg
			}
		}
	}
	return packages, nil
}

func renderTable(ctx api.Context, list []cosmos.Package) {
	table := cli.NewTable(ctx.Out(), []string{"NAME", "VERSION", "CERTIFIED", "APP", "COMMAND", "DESCRIPTION"})
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
}

func filterPackages(pkgs map[id]cosmos.Package, filter func(app string) bool) []cosmos.Package {
	filteredList := make([]cosmos.Package, 0, len(pkgs))
	for _, pkg := range pkgs {
		if filter(pkg.Name) {
			filteredList = append(filteredList, pkg)
			continue
		}
		for _, app := range pkg.Apps {
			if filter(app) {
				filteredList = append(filteredList, pkg)
				break
			}
		}
	}

	sort.Sort(packages(filteredList))
	return filteredList
}

type id string

func identifier(p cosmos.Package) id {
	return id(p.Name + p.Version)
}

type packages []cosmos.Package

func (p packages) Len() int           { return len(p) }
func (p packages) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p packages) Less(i, j int) bool { return identifier(p[i]) < identifier(p[j]) }

func appendIfMissing(slice []string, s ...string) []string {
LOOP:
	for _, i := range s {
		for _, ele := range slice {
			if ele == i {
				continue LOOP
			}
		}
		slice = append(slice, i)
	}
	return slice
}
