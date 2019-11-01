package marathon

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/python"
	"github.com/spf13/cobra"
)

func newCmdMarathonPing(ctx api.Context) *cobra.Command {
	var once bool

	cmd := &cobra.Command{
		Use:   "ping",
		Short: "Ensure Marathon is up and responding.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return python.InvokePythonCLI(ctx)
		},
	}

	// TODO: the Python help output doesn't give this a description also it already only pings once so what does this do?
	cmd.Flags().BoolVar(&once, "once", false, "")

	return cmd
}
