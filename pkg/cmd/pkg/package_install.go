package pkg

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dcos/client-go/dcos"
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/plugin"
	"github.com/dcos/dcos-cli/pkg/prompt"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/dcos/dcos-core-cli/pkg/cosmos"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
)

func newCmdPackageInstall(ctx api.Context) *cobra.Command {
	var appOnly, cliOnly, yes bool
	var appID, optionsPath, version string

	cmd := &cobra.Command{
		Use:   "install <package-name>",
		Short: "Install a package",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			packageName := args[0]

			if cliOnly == appOnly {
				cliOnly = true
				appOnly = true
			}

			c, err := cosmos.NewClient(ctx, pluginutil.HTTPClient(""))
			if err != nil {
				return err
			}

			description, err := c.PackageDescribe(packageName, version)
			if err != nil {
				return err
			}
			pkg := description.Package

			link := "https://mesosphere.com/catalog-terms-conditions/#certified-services"
			if !pkg.Selected {
				fmt.Fprint(ctx.Out(), "This is a Community service. "+
					"Community services are not tested for production environments. "+
					"There may be bugs, incomplete features, incorrect documentation, or other discrepancies.\n")
				link = "https://mesosphere.com/catalog-terms-conditions/#community-services"
			}
			fmt.Fprintf(ctx.Out(), "By Deploying, you agree to the Terms and Conditions %s\n", link)
			if appOnly && pkg.PreInstallNotes != "" {
				_, err := fmt.Fprintf(ctx.Out(), "%s\n", pkg.PreInstallNotes)
				if err != nil {
					return err
				}
			}

			if !yes {
				prompter := prompt.New(ctx.Input(), ctx.Out())
				err = prompter.Confirm("Continue installing? [yes/no] ", "Yes")
				if err != nil {
					return err
				}
			}

			if appOnly && description.Package.Marathon.V2AppMustacheTemplate != "" {
				fmt.Fprintf(ctx.Out(), "Installing Marathon app for package [%s] version [%s]\n", packageName, pkg.Version)
				err := c.PackageInstalls(appID, packageName, pkg.Version, optionsPath)
				if err != nil {
					return err
				}
			}

			fmt.Fprintf(ctx.Out(), "%s\n", pkg.PostInstallNotes)
			if cliOnly && !isEmptyCli(pkg.Resource.Cli) {
				fmt.Fprintf(ctx.Out(), "Installing CLI subcommand for package [%s] version [%s]\n", packageName, pkg.Version)
				pluginInfo, err := cosmos.CLIPluginInfo(description.Package, pluginutil.HTTPClient("").BaseURL())
				if err != nil {
					return err
				}

				var checksum plugin.Checksum
				for _, contentHash := range pluginInfo.ContentHash {
					switch contentHash.Algo {
					case dcos.SHA256:
						checksum.Hasher = sha256.New()
						checksum.Value = contentHash.Value
					}
				}

				cluster, err := ctx.Cluster()
				if err != nil {
					return err
				}

				err = ctx.PluginManager(cluster).Install(pluginInfo.Url, &plugin.InstallOpts{
					Name:     packageName,
					Update:   true,
					Checksum: checksum,
					PostInstall: func(fs afero.Fs, pluginDir string) error {
						pkgInfoFilepath := filepath.Join(pluginDir, "package.json")
						pkgInfoFile, err := fs.OpenFile(pkgInfoFilepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
						if err != nil {
							return err
						}
						defer pkgInfoFile.Close()
						return json.NewEncoder(pkgInfoFile).Encode(description.Package)
					},
				})
				if err != nil {
					return err
				}
				plugin, err := ctx.PluginManager(cluster).Plugin(packageName)
				if err != nil {
					return err
				}

				plural := ""
				if len(plugin.Commands) > 1 {
					plural = "s"
				}
				cmds := make([]string, 0, len(plugin.Commands))
				for _, c := range plugin.Commands {
					cmds = append(cmds, c.Name)
				}

				fmt.Fprintf(ctx.Out(), "New command%s available: dcos %s\n", plural, strings.Join(cmds, ", "))
			}

			return nil
		},
	}
	cmd.Flags().BoolVar(&appOnly, "app", false, "Application only")
	cmd.Flags().StringVar(&appID, "app-id", "", "The application ID")
	cmd.Flags().BoolVar(&cliOnly, "cli", false, "Command line only")
	cmd.Flags().StringVar(&optionsPath, "options", "", "Path to a JSON file that contains customized package installation options")
	cmd.Flags().StringVar(&version, "package-version", "", "The package version")
	cmd.Flags().BoolVar(&yes, "yes", false, "Disable interactive mode and assume “yes” is the answer to all prompts")
	return cmd
}
