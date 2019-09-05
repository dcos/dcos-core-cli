package pkg

import (
	"errors"
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/prompt"
	"github.com/dcos/dcos-core-cli/pkg/cosmos"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

func newCmdPackageUninstall(ctx api.Context) *cobra.Command {
	var all, appOnly, cliOnly, yes bool
	var appID string

	cmd := &cobra.Command{
		Use:   "uninstall <package-name>",
		Short: "Uninstall a package",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			packageName := args[0]

			// TODO: check flag combination compatibilities
			if cliOnly && appOnly {
				return errors.New("--cli and --app cannot be used together")
			}

			if cliOnly {
				return removeCLI(ctx, packageName)
			}

			c, err := cosmos.NewClient(ctx, pluginutil.HTTPClient(""))
			if err != nil {
				return err
			}

			installed, err := packageInstalled(c, packageName)
			if err != nil {
				return err
			}
			if !installed {
				return fmt.Errorf("Package '%s' is not installed", packageName)
			}

			if !uninstallConfirmed(ctx, packageName) {
				return errors.New("Cancelling uninstall.")
			}

			err = c.PackageUninstall(packageName, all, appID)
			if err != nil {
				return err
			}

			if !appOnly {
				err = removeCLI(ctx, packageName)
			}

			return err
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "All applications")
	cmd.Flags().BoolVar(&appOnly, "app", false, "Application only")
	cmd.Flags().StringVar(&appID, "app-id", "", "The application ID")
	cmd.Flags().BoolVar(&cliOnly, "cli", false, "Command line only")
	cmd.Flags().BoolVar(&yes, "yes", false, "Disable interactive mode and assume “yes” is the answer to all prompts")
	return cmd
}

func removeCLI(ctx api.Context, packageName string) error {
	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}

	manager := ctx.PluginManager(cluster)
	return manager.Remove(packageName)
}

func packageInstalled(c *cosmos.Client, packageName string) (bool, error) {
	packages, err := c.PackageList()
	if err != nil {
		return false, err
	}

	for _, p := range *packages {
		if packageName == p.Name {
			return true, nil
		}
	}

	return false, nil
}

func uninstallConfirmed(ctx api.Context, expected string) bool {
	p := prompt.New(ctx.Input(), ctx.Out())
	for i := 0; i < 3; i++ {
		response := p.Input("Please type the name of the service to confirm: ")
		if response == expected {
			return true
		}
		fmt.Fprintf(ctx.Out(), "Expected '%s'. You supplied '%s'.", expected, response)
	}
	return false
}
