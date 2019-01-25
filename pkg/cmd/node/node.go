package node

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/mesos"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

// NewCommand creates the `core node` subcommand.
func NewCommand(ctx api.Context) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "node",
		Short: "Display DC/OS node information",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, ok := ctx.EnvLookup(cli.EnvStrictDeprecations)
			if !ok {
				ctx.Deprecated("Getting the list of nodes from `dcos node` is deprecated. Please use `dcos node list`.")
				return newCmdNodeList(ctx).RunE(cmd, args)
			}
			return cmd.Help()
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")

	cmd.AddCommand(
		newCmdNodeDecommission(ctx),
		newCmdNodeList(ctx),
		newCmdNodeListComponents(ctx),
	)
	return cmd
}

func mesosClient() *mesos.Client {
	return mesos.NewClient(pluginutil.HTTPClient(""))
}
