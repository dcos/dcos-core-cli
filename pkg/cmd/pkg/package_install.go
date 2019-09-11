package pkg

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/prompt"
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

			link := "https://mesosphere.com/catalog-terms-conditions/#community-services"
			if description.Package.Selected {
				link = "https://mesosphere.com/catalog-terms-conditions/#certified-services"
			}
			_, err = fmt.Fprintf(ctx.Out(), "By Deploying, you agree to the Terms and Conditions %s\n", link)
			if err != nil {
				return err
			}

			if appOnly && description.Package.PreInstallNotes != "" {
				_, err := fmt.Fprintf(ctx.Out(), "%s\n", description.Package.PreInstallNotes)
				if err != nil {
					return err
				}
			}

			prompter := prompt.New(ctx.Input(), ctx.Out())
			err = prompter.Confirm("Continue installing? [yes/no] ", "No")
			if err != nil {
				return err
			}

			if appOnly {
				_, err = fmt.Fprintf(ctx.Out(), "Installing Marathon app for package [%s] version [%s]\n", packageName, description.Package.Version)
				if err != nil {
					return err
				}

				postInstallNotes, err := c.PackageInstalls(appID, packageName, description.Package.Version, optionsPath)
				if err != nil {
					return err
				}
				_, err = fmt.Fprintf(ctx.Out(), "%s\n", postInstallNotes)
				if err != nil {
					return err
				}
			}

			return invokePythonCLI(ctx)
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
