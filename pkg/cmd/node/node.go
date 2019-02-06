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
	cmd := &cobra.Command{
		Use:   "node",
		Short: "Display DC/OS node information",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, ok := ctx.EnvLookup(cli.EnvStrictDeprecations)
			if !ok {
				ctx.Deprecated("Getting the list of nodes from `dcos node` is deprecated. Please use `dcos node list`.")
				listCmd := newCmdNodeList(ctx)
				// Execute by default would use os.Args[1:], which is everything after `dcos ...`.
				// We need all command line arguments after `dcos node ...`.
				listCmd.SetArgs(ctx.Args()[2:])
				return listCmd.Execute()
			}
			return cmd.Help()
		},
	}
	cmd.Flags().Bool("json", false, "Print in json format")
	cmd.Flags().StringArray("field", nil, "Name of extra field to include in the output of `dcos node`. Can be repeated multiple times to add several fields.")

	cmd.AddCommand(
		newCmdNodeDecommission(ctx),
		newCmdNodeDNS(ctx),
		newCmdNodeList(ctx),
		newCmdNodeListComponents(ctx),
		newCmdNodeMetrics(ctx),
	)
	return cmd
}

func diagnosticsClient() *diagnostics.Client {
	return diagnostics.NewClient(pluginutil.HTTPClient(""))
}

func mesosClient() *mesos.Client {
	return mesos.NewClient(pluginutil.HTTPClient(""))
}
