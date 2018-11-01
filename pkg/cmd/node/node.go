package node

import (
	"crypto/tls"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-cli/pkg/httpclient"
	"github.com/dcos/dcos-core-cli/pkg/mesos"
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
		newCmdNodeList(ctx),
	)

	return cmd
}

func mesosClient(ctx api.Context) (*mesos.Client, error) {
	cluster, err := ctx.Cluster()
	if err != nil {
		return nil, err
	}

	baseURL, _ := cluster.Config().Get("core.dcos_url").(string)
	if baseURL == "" {
		baseURL = cluster.URL()
	}

	return mesos.NewClient(
		httpclient.New(
			baseURL,
			httpclient.Logger(ctx.Logger()),
			httpclient.ACSToken(cluster.ACSToken()),
			httpclient.Timeout(cluster.Timeout()),
			httpclient.TLS(&tls.Config{
				InsecureSkipVerify: cluster.TLS().Insecure,
				RootCAs:            cluster.TLS().RootCAs,
			}),
		),
	), nil

}
