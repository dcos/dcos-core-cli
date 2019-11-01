package marathon

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/python"
	"github.com/spf13/cobra"
)

func newCmdMarathonDelay(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delay",
		Short: "Control Marathon deployment delay.",
	}

	cmd.AddCommand(
		newCmdMarathonDelayReset(ctx),
	)
	return cmd
}

func newCmdMarathonDelayReset(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset the current delay (if any) of the application.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return python.InvokePythonCLI(ctx)
		},
	}
	return cmd
}
