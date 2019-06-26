package node

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/diagnostics"
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
			if len(args) == 0 {
				_, ok := ctx.EnvLookup(cli.EnvStrictDeprecations)
				if !ok {
					ctx.Deprecated("Getting the list of nodes from `dcos node` is deprecated. Please use `dcos node list`.")
					listCmd := newCmdNodeList(ctx)
					// Execute by default would use os.Args[1:], which is everything after `dcos ...`.
					// We need all command line arguments after `dcos node ...`.
					listCmd.SetArgs(ctx.Args()[2:])
					listCmd.SilenceErrors = true
					listCmd.SilenceUsage = true
					return listCmd.Execute()
				}
				return cmd.Help()
			}
			fmt.Fprintln(ctx.ErrOut(), cmd.UsageString())
			return fmt.Errorf("unknown command %s", args[0])
		},
	}
	cmd.Flags().Bool("json", false, "Print in json format")
	cmd.Flags().StringArray("field", nil, "Name of extra field to include in the output of `dcos node`. Can be repeated multiple times to add several fields.")

	cmd.AddCommand(
		newCmdNodeDecommission(ctx),
		newCmdNodeDiagnostics(ctx),
		newCmdNodeDNS(ctx),
		newCmdNodeList(ctx),
		newCmdNodeListComponents(ctx),
		newCmdNodeLog(ctx),
		newCmdNodeMetrics(ctx),
		newCmdNodeSSH(ctx),
	)
	return cmd
}

func diagnosticsClient() *diagnostics.Client {
	return diagnostics.NewClient(pluginutil.HTTPClient(""))
}

// mesosDNSClient returns a client with a`baseURL` to communicate with Mesos-DNS.
func mesosDNSClient() *mesos.Client {
	return mesos.NewClient(pluginutil.HTTPClient(""))
}

// mesosClient returns a client with a `baseURL` to communicate with Mesos.
func mesosClient(ctx api.Context) (*mesos.Client, error) {
	cluster, err := ctx.Cluster()
	if err != nil {
		return nil, err
	}
	baseURL, _ := cluster.Config().Get("core.mesos_master_url").(string)
	if baseURL == "" {
		baseURL = cluster.URL() + "/mesos"
	}
	return mesos.NewClient(pluginutil.HTTPClient(baseURL)), nil
}
