package task

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/python"
	"github.com/spf13/cobra"
)

func newCmdMarathonTaskStop(ctx api.Context) *cobra.Command {
	var wipe bool

	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop a task.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return python.InvokePythonCLI(ctx)
		},
	}

	cmd.Flags().BoolVar(&wipe, "wipe", false, "Wipe persistent data.")

	return cmd
}
