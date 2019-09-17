package pkg

import (
	"errors"
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cosmos"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

const (
	removeAllWarningTemplate = `WARNING: This action cannot be undone. This will uninstall ` +
		`all instances of the [%s] package, and delete all data for all %s instances.`

	removeAppWarningTemplate = `WARNING: This action cannot be undone. This will uninstall ` +
		`[%s] and delete all of its persistent data (logs, configurations, database artifacts, everything).`
)

func newCmdPackageUninstall(ctx api.Context) *cobra.Command {
	var all, appOnly, cliOnly, yes bool
	var appID string

	cmd := &cobra.Command{
		Use:   "uninstall <package-name>",
		Short: "Uninstall a package",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if cliOnly && (appOnly || appID != "" || all) {
				return errors.New("--cli cannot be used with --app, --app-id, or --all")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			packageName := args[0]

			if cliOnly {
				return removeCLI(ctx, packageName)
			}

			c, err := cosmos.NewClient(ctx, pluginutil.HTTPClient(""))
			if err != nil {
				return err
			}

			installed, err := packageInstallCount(c, packageName)
			if err != nil {
				return err
			}
			if installed == 0 {
				return fmt.Errorf("Package [%s] is not installed", packageName)
			}

			name := packageName
			removingApp := appID != ""
			if removingApp {
				name = appID
			}

			if !yes {
				fmt.Fprintln(ctx.ErrOut(), warningMessage(name, all))
			}
			promptMsg := promptMessage(name, removingApp, all)

			expected := name
			if all {
				expected = fmt.Sprintf("uninstall all %s", name)
			}
			if !uninstallConfirmed(ctx, expected, promptMsg, yes) {
				return errors.New("cancelling uninstall")
			}

			resp, err := c.PackageUninstall(packageName, all, appID)
			if err != nil {
				return err
			}
			for _, r := range resp.Results {
				fmt.Fprintf(ctx.ErrOut(), "Uninstalled package [%s] version [%s]\n", r.PackageName, r.PackageVersion)
				if r.PostUninstallNotes != "" {
					fmt.Fprintln(ctx.ErrOut(), r.PostUninstallNotes)
				}
			}
			// only remove the CLI if there aren't any other of the same package remaining
			// so only 1 was installed at the start
			// calling this with --app when there are multiple apps will not result in
			// all of those being removed so we should only have to check if there was 1
			// instance installed already
			if !appOnly && (all || installed == 1) {
				removeCLI(ctx, packageName)
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

func packageInstallCount(c *cosmos.Client, packageName string) (int, error) {
	packages, err := c.PackageList()
	if err != nil {
		return 0, err
	}

	for _, p := range packages {
		if packageName == p.Name {
			return len(p.Apps), nil
		}
	}

	return 0, nil
}

func warningMessage(name string, all bool) string {
	if all {
		return fmt.Sprintf(removeAllWarningTemplate, name, name)
	}
	return fmt.Sprintf(removeAppWarningTemplate, name)
}

func promptMessage(name string, appID bool, all bool) string {
	if all {
		return fmt.Sprintf("Please type 'uninstall all %s': ", name)
	}
	if appID {
		return "Please type the full name of the app ID to confirm: "
	}
	// if not all or appID, default to removing the package (i.e. --app but not only the app)
	return "Please type the name of the service to confirm: "
}

func uninstallConfirmed(ctx api.Context, expected string, promptMsg string, skip bool) bool {
	if skip {
		return true
	}

	p := ctx.Prompt()

	for i := 0; i < 3; i++ {
		response := p.Input(promptMsg)
		if response == expected {
			return true
		}
		fmt.Fprintf(ctx.ErrOut(), "Expected '%s'. You supplied '%s'.\n", expected, response)
	}
	return false
}
