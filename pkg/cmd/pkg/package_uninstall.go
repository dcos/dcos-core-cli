package pkg

import (
	"github.com/dcos/dcos-cli/api"
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
			return invokePythonCLI(ctx)
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "All applications")
	cmd.Flags().BoolVar(&appOnly, "app", false, "Application only")
	cmd.Flags().StringVar(&appID, "app-id", "", "The application ID")
	cmd.Flags().BoolVar(&cliOnly, "cli", false, "Command line only")
	cmd.Flags().BoolVar(&yes, "yes", false, "Disable interactive mode and assume “yes” is the answer to all prompts")
	return cmd
}
