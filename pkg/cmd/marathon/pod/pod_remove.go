package pod

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/python"
	"github.com/spf13/cobra"
)

func newCmdMarathonPodRemove(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a pod.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return python.InvokePythonCLI(ctx)
		},
	}
	return cmd
}
