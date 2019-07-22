package quota

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/mesos"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

// NewCommand creates the `dcos quota` subcommand.
func NewCommand(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quota",
		Short: "Manage DC/OS quotas",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			fmt.Fprintln(ctx.ErrOut(), cmd.UsageString())
			return fmt.Errorf("unknown command %s", args[0])
		},
	}

	cmd.AddCommand(
		newCmdQuotaCreate(ctx),
		newCmdQuotaDelete(ctx),
		newCmdQuotaGet(ctx),
		newCmdQuotaUpdate(ctx),
	)

	return cmd
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
