package pkg

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/dcos/client-go/dcos"
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cosmos"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

func newCmdPackageDescribe(ctx api.Context) *cobra.Command {
	var allVersions, appOnly, cliOnly, config, render bool
	var appID, optionsPath, version string

	cmd := &cobra.Command{
		Use:   "describe <package-name>",
		Short: "Get specific details for packages",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			packageName := args[0]

			c, err := cosmos.NewClient(ctx, pluginutil.HTTPClient(""))
			if err != nil {
				return err
			}
			enc := json.NewEncoder(ctx.Out())
			enc.SetIndent("", "  ")

			if allVersions {
				versions, err := c.PackageListVersions(packageName)
				if err != nil {
					return err
				}
				return enc.Encode(versions)

			}

			description, err := c.PackageDescribe(packageName, version)
			if err != nil {
				if version == "" {
					return fmt.Errorf("package [%s] not found. Find the correct package name using 'dcos package search': %s",
						packageName, err)
				}
				return fmt.Errorf("version [%s] of package [%s] not found: %s", version, packageName, err)
			}

			if appOnly {
				// If the user supplied template options, they definitely want to render the template
				if render || optionsPath != "" {
					template, err := c.PackageRender(appID, packageName, version, optionsPath)
					if err != nil {
						return err
					}
					return enc.Encode(template)
				}

				template, err := base64.StdEncoding.DecodeString(description.Package.Marathon.V2AppMustacheTemplate)
				if err != nil {
					return err
				}
				_, err = ctx.Out().Write(template)
				return err
			}

			if cliOnly {
				if !isEmptyCli(description.Package.Resource.Cli) {
					return enc.Encode(description.Package.Resource.Cli)
				}
				if !isEmptyCommand(description.Package.Command) {
					return enc.Encode(description.Package.Command)
				}
				return nil
			}

			if config {
				return enc.Encode(description.Package.Config)
			}

			return enc.Encode(description)
		},
	}
	cmd.Flags().BoolVar(&appOnly, "app", false, "Application only")
	cmd.Flags().StringVar(&appID, "app-id", "", "The application ID")
	cmd.Flags().BoolVar(&cliOnly, "cli", false, "Command line only")
	cmd.Flags().BoolVar(&config, "config", false, "Print the configurable properties of the `marathon.json` file")
	cmd.Flags().StringVar(&optionsPath, "options", "", "Path to a JSON file that contains customized package installation options")
	cmd.Flags().StringVar(&version, "package-version", "", "The package version")
	cmd.Flags().BoolVar(&allVersions, "package-versions", false, " Displays all versions for this package")
	cmd.Flags().BoolVar(&render, "render", false, "Collate the `marathon.json` package template with values from the `config.json` and `--options`")
	return cmd
}

func isEmptyCommand(cmd dcos.CosmosPackageCommand) bool {
	return cmd.Name == "" && len(cmd.Pip) == 0
}

func isEmptyCli(cli dcos.CosmosPackageResourceCli) bool {
	return isEmptyBinary(cli.Binaries.Darwin) &&
		isEmptyBinary(cli.Binaries.Linux) &&
		isEmptyBinary(cli.Binaries.Windows)
}

func isEmptyBinary(binary dcos.CosmosPackageResourceCliOsBinaries) bool {
	x := binary.X8664
	return len(x.ContentHash) == 0 && x.Kind == "" && x.Url == ""
}
