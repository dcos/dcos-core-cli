package task

import (
	"os"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/mesos"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli/httpagent"
	"github.com/spf13/cobra"
)

func newCmdTaskAttach(ctx api.Context) *cobra.Command {
	var noStdin bool

	cmd := &cobra.Command{
		Use: "attach",
		RunE: func(cmd *cobra.Command, args []string) error {
			task, err := findTask(ctx, args[0])
			if err != nil {
				return err
			}

			httpClient, err := mesosHTTPClient(ctx, task.AgentID.Value)
			if err != nil {
				return err
			}

			containerID := task.Statuses[0].ContainerStatus.ContainerID

			taskIO, err := mesos.NewTaskIO(*containerID, mesos.TaskIOOpts{
				Stdin:       ctx.Input(),
				Stdout:      ctx.Out(),
				Stderr:      ctx.ErrOut(),
				Interactive: !noStdin,
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
