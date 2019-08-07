package pkg

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdPackageInstall(ctx api.Context) *cobra.Command {
	var appOnly, cliOnly, global, yes bool
	var appID, optionsPath, version string

	cmd := &cobra.Command{
		Use:   "install <package-name>",
		Short: "Install a package",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return invokePythonCLI(ctx)
		},
	}
	cmd.Flags().BoolVar(&appOnly, "app", false, "Application only")
	cmd.Flags().StringVar(&appID, "app-id", "", "The application ID")
	cmd.Flags().BoolVar(&cliOnly, "cli", false, "Command line only")
	cmd.Flags().BoolVar(&global, "global", false, "Install a subcommand for all configured clusters")
	cmd.Flags().StringVar(&optionsPath, "options", "", "Path to a JSON file that contains customized package installation options")
	cmd.Flags().StringVar(&version, "package-version", "", "The package version")
	cmd.Flags().BoolVar(&yes, "yes", false, "Disable interactive mode and assume “yes” is the answer to all prompts")
	return cmd
}
