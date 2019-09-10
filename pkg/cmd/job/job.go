package job

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/metronome"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var errTerminalInput = errors.New("input from the terminal is not accepted")

// NewCommand creates the `core job` subcommand.
func NewCommand(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Deploy and manage jobs in DC/OS",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			fmt.Fprintln(ctx.ErrOut(), cmd.UsageString())
			return fmt.Errorf("unknown command %s", args[0])
		},
	}

	cmd.AddCommand(
		newCmdJobAdd(ctx),
		newCmdJobHistory(ctx),
		newCmdJobKill(ctx),
		newCmdJobList(ctx),
		newCmdJobQueue(ctx),
		newCmdJobRemove(ctx),
		newCmdJobRun(ctx),
		newCmdJobSchedule(ctx),
		newCmdJobShow(ctx),
		newCmdJobUpdate(ctx),
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
	return metronome.NewClient(pluginutil.HTTPClient(baseURL), pluginutil.Logger()), nil
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
			return nil, errTerminalInput
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
