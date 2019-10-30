package task

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/python"
	"github.com/spf13/cobra"
)

func newCmdMarathonTaskKill(ctx api.Context) *cobra.Command {
	var scale bool
	var wipe bool
	var json bool

	cmd := &cobra.Command{
		Use:   "kill",
		Short: "Kill one or more tasks.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return python.InvokePythonCLI(ctx)
		},
	}

	cmd.Flags().BoolVar(&scale, "scale", false, "Scale the app down after performing the the operation.")
	cmd.Flags().BoolVar(&wipe, "wipe", false, "Wipe persistent data.")
	cmd.Flags().BoolVar(&json, "json", false, "Print JSON-formatted data.")

	return cmd
}
