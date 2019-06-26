package task

import (
	"os"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/mesos"
	mesosgo "github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli/httpagent"
	"github.com/spf13/cobra"
)

func newCmdTaskAttach(ctx api.Context) *cobra.Command {
	var noStdin bool

	cmd := &cobra.Command{
		Use:   "attach <task>",
		Short: "Attach the CLI to the stdio of an already running task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filters := taskFilters{
				Active: true,
				ID:     args[0],
			}

			task, err := findTask(ctx, filters)
			if err != nil {
				return err
			}

			httpClient, err := mesosHTTPClient(ctx, task.SlaveID)
			if err != nil {
				return err
			}

			containerID := mesosgo.ContainerID{
				Value: task.Statuses[0].ContainerStatus.ContainerID.Value,
			}

			taskIO, err := mesos.NewTaskIO(containerID, mesos.TaskIOOpts{
				Stdin:       ctx.Input(),
				Stdout:      ctx.Out(),
				Stderr:      ctx.ErrOut(),
				Interactive: !noStdin,
				TTY:         true,
				Sender:      httpagent.NewSender(httpClient.Send),
			})

			if err != nil {
				return err
			}

			exitCode, err := taskIO.Attach()
			if err != nil {
				return err
			}
			os.Exit(exitCode)
			return nil
		},
	}

	cmd.Flags().BoolVar(&noStdin, "no-stdin", false, "Don't attach the stdin of the CLI to the task")

	return cmd
}
