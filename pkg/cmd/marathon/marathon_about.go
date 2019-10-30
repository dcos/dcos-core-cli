package marathon

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/python"
	"github.com/spf13/cobra"
)

func newCmdMarathonAbout(ctx api.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "about",
		Short: "Print info.json for DC/OS Marathon.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return python.InvokePythonCLI(ctx)
		},
	}
}
