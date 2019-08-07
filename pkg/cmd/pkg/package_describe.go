package pkg

import (
	"github.com/dcos/dcos-cli/api"
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
			return invokePythonCLI(ctx)
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
