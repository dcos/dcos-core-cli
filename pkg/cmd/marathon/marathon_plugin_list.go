package marathon

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/python"
	"github.com/spf13/cobra"
)

func newCmdMarathonPlugin(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage Marathon plugins.",
	}

	cmd.AddCommand(
		newCmdMarathonPluginList(ctx),
	)

	return cmd
}

func newCmdMarathonPluginList(ctx api.Context) *cobra.Command {
	var json bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List plugins.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return python.InvokePythonCLI(ctx)
		},
	}

	cmd.Flags().BoolVar(&json, "json", false, "Print JSON-formatted data.")

	return cmd
}
