package job

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/httpclient"
	"github.com/dcos/dcos-core-cli/pkg/metronome"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

// NewCommand creates the `core job` subcommand.
func NewCommand(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Deploying and managing jobs in DC/OS",
	}

	cmd.AddCommand(
		newCmdJobAdd(ctx),
		newCmdJobList(ctx),
		newCmdJobUpdate(ctx),
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

func parseJSONJob(r io.Reader) (*metronome.Job, error) {
	jsonBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var job metronome.Job
	if err := json.Unmarshal(jsonBytes, &job); err != nil {
		return nil, err
	}

	return &job, nil
}

func inputReader(ctx api.Context, args []string) (io.Reader, error) {
	switch len(args) {
	case 0:
		input, _ := ctx.Input().(*os.File)
		if terminal.IsTerminal(int(input.Fd())) {
			return nil, fmt.Errorf("input from the terminal is not accepted")
		}
		return ctx.Input(), nil
	case 1:
		reader, err := ctx.Fs().Open(args[0])
		if err != nil {
			return nil, err
		}
		return reader, nil
	default:
		return nil, fmt.Errorf("input must be from stdin or file")
	}
}
