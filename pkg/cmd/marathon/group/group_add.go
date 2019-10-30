package group

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/python"
	"github.com/spf13/cobra"
)

func newCmdMarathonGroupAdd(ctx api.Context) *cobra.Command {
	var groupID string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a group.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return python.InvokePythonCLI(ctx)
		},
	}

	cmd.Flags().StringVar(&groupID, "id", "", "The group ID to add.")

	return cmd
}
