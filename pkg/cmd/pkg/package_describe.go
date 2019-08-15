package pkg

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cosmos"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

func newCmdPackageDescribe(ctx api.Context) *cobra.Command {
	var versions, appOnly, cliOnly, configOnly, render bool
	var appID, optionsPath, version string

	cmd := &cobra.Command{
		Use:   "describe <package-name>",
		Short: "Get specific details for packages",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if optionsPath != "" {
				if _, err := os.Stat(optionsPath); os.IsNotExist(err) {
					return fmt.Errorf("path '%s' does not exist", optionsPath)
				}
				render = true
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := cosmos.NewClient(ctx, pluginutil.HTTPClient(""))
			if err != nil {
				return err
			}

			if versions {
				list, err := c.PackageListVersions(args[0])
				if err != nil {
					return err
				}

				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(list)
			}

			if appOnly && render {
				desc, err := c.PackageRender(appID, args[0], version, optionsPath)
				if err != nil {
					return err
				}

				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(desc)
			}

			desc, err := c.PackageDescribe(args[0], version)
			if err != nil {
				return err
			}

			if appOnly {
				marathon, err := base64.StdEncoding.DecodeString(desc.Package.Marathon.V2AppMustacheTemplate)
				if err != nil {
					return err
				}

				fmt.Fprintf(ctx.Out(), string(marathon))
				return nil
			}

			if cliOnly {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(desc.Package.Resource.Cli)
			}

			if configOnly {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(desc.Package.Config)
			}

			if appOnly {
				if render {

				}

			}
			enc := json.NewEncoder(ctx.Out())
			enc.SetIndent("", "    ")
			return enc.Encode(desc)
		},
	}
	cmd.Flags().BoolVar(&appOnly, "app", false, "Application only")
	cmd.Flags().StringVar(&appID, "app-id", "", "The application ID used if rendering the package")
	cmd.Flags().BoolVar(&cliOnly, "cli", false, "Command line only")
	cmd.Flags().BoolVar(&configOnly, "config", false, "Print the configurable properties of the `marathon.json` file")
	cmd.Flags().StringVar(&optionsPath, "options", "", "Path to a JSON file that contains customized package installation options")
	cmd.Flags().StringVar(&version, "package-version", "", "The package version")
	cmd.Flags().BoolVar(&versions, "package-versions", false, " Displays all versions for this package")
	cmd.Flags().BoolVar(&render, "render", false, "Collate the `marathon.json` package template with values from the `config.json` and `--options`")
	return cmd
}
