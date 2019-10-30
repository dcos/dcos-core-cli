package group

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/python"
	"github.com/spf13/cobra"
)

func newCmdMarathonGroupScale(ctx api.Context) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "scale",
		Short: "Scale a group.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return python.InvokePythonCLI(ctx)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Disable checks in Marathon during updates.")

	return cmd
}
