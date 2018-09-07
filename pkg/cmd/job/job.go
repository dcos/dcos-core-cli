package job

import (
	"crypto/tls"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/httpclient"
	"github.com/dcos/dcos-core-cli/pkg/metronome"
	"github.com/spf13/cobra"
)

// NewCommand creates the `core job` subcommand.
func NewCommand(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Deploying and managing jobs in DC/OS",
	}

	cmd.AddCommand(
		newCmdJobList(ctx),
		newCmdJobRemove(ctx),
		newCmdJobRun(ctx),
	)

	return cmd
}

func metronomeClient(ctx api.Context) (*metronome.Client, error) {
	cluster, err := ctx.Cluster()
	if err != nil {
		return nil, err
	}

	baseURL, _ := cluster.Config().Get("job.url").(string)
	if baseURL == "" {
		baseURL = cluster.URL() + "/service/metronome"
	}

	return metronome.NewClient(httpclient.New(
		baseURL,
		httpclient.Logger(ctx.Logger()),
		httpclient.ACSToken(cluster.ACSToken()),
		httpclient.Timeout(cluster.Timeout()),
		httpclient.TLS(&tls.Config{
			InsecureSkipVerify: cluster.TLS().Insecure,
			RootCAs:            cluster.TLS().RootCAs,
		}),
	)), nil

}
