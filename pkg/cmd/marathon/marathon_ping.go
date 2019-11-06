package marathon

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/marathon"
	"github.com/dcos/dcos-core-cli/pkg/mesos"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

func newCmdMarathonPing(ctx api.Context) *cobra.Command {
	var once bool

	cmd := &cobra.Command{
		Use:   "ping",
		Short: "Ensure Marathon is up and responding.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := marathon.NewClient(ctx)
			if err != nil {
				return err
			}

			masters, err := marathonPing(ctx, *client, once)
			if err != nil {
				return err
			}

			// If no err, a pong has been received.
			fmt.Fprintf(ctx.Out(), "Marathon ping response[%dx]: \"pong\"\n", masters)
			return nil
		},
	}

	// TODO: the Python help output doesn't give this a description also it already only pings once so what does this do?
	cmd.Flags().BoolVar(&once, "once", false, "")

	return cmd
}

func marathonPing(ctx api.Context, client marathon.Client, once bool) (int, error) {
	numberOfMasters := 1
	if !once {
		cluster, err := ctx.Cluster()
		if err != nil {
			return 0, err
		}
		mesosClient := mesos.NewClient(pluginutil.HTTPClient(cluster.URL()))
		masters, err := mesosClient.Masters()
		if err != nil {
			return 0, err
		}
		numberOfMasters = len(masters)
	}

	for i := 1; i <= numberOfMasters; i++ {
		_, err := client.API.Ping()
		if err != nil {
			return 0, fmt.Errorf("unable to ping leading Marathon master %d time(s)", i)
		}
	}
	return numberOfMasters, nil
}
