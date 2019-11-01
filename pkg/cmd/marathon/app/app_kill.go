package app

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/python"
	"github.com/spf13/cobra"
)

func newCmdMarathonAppKill(ctx api.Context) *cobra.Command {
	var scale bool
	var host string

	cmd := &cobra.Command{
		Use:   "kill",
		Short: "Kill a running application instance.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return python.InvokePythonCLI(ctx)
		},
	}

	cmd.Flags().BoolVar(&scale, "scale", false, "Scale the app down after performing the the operation.")
	cmd.Flags().StringVar(&host, "host", "", "The hostname that is running app.")

	return cmd
}
