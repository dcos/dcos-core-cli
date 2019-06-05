package task

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/mesos"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	lib "github.com/mesos/mesos-go/api/v1/lib"
	"github.com/spf13/cobra"
)

func newCmdTaskAttach(ctx api.Context) *cobra.Command {
	// all and completed are useless but we keep them for retrocompatibility.
	var noStdin bool

	cmd := &cobra.Command{
		Use:   "attach <task>",
		Short: " Attach the CLI to the stdio of an already running task (<task>). To detach type the sequence CTRL-p CTRL-q.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mesosClient, err := mesos.NewClientWithContext(ctx)
			if err != nil {
				return err
			}

			allTasks, err := mesosClient.Tasks()
			if err != nil {
				return err
			}

			var tasks []lib.Task
			for _, t := range allTasks {
				if strings.Contains(t.TaskID.Value, args[0]) {
					tasks = append(tasks, t)
				}
			}

			if len(tasks) == 0 {
				return fmt.Errorf("no task ID found containing '%s'", args[0])
			} else if len(tasks) > 1 {
				return fmt.Errorf("found more than one task with the same name: %v", tasks)
			}

			task := tasks[0]
			if task.Container.Type.String() != "MESOS" {
				return fmt.Errorf("this command is only supported for tasks launched by the Universal Container Runtime (UCR)")
			}

			// Random string.
			rand.Seed(time.Now().UnixNano())
			chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZÅÄÖ" +
				"abcdefghijklmnopqrstuvwxyzåäö" +
				"0123456789")
			length := 8
			var b strings.Builder
			for i := 0; i < length; i++ {
				b.WriteRune(chars[rand.Intn(len(chars))])
			}

			// In the Python implementation we take the first status but the last one makes more sense.
			cid := task.Statuses[len(task.Statuses)-1].ContainerStatus.ContainerID
			nestedContainer := lib.ContainerID{
				Value:  b.String(),
				Parent: cid,
			}

			client := mesos.NewClient(pluginutil.HTTPClient(""))
			// Creates the nested container where commands will be run.
			err = client.LaunchNestedContainer(task.AgentID.Value, nestedContainer)
			if err != nil {
				return err
			}

			return nil
		},
	}
	cmd.Flags().BoolVar(&noStdin, "--no-stdin", false, "Do not attach the stdin of the CLI to the task")
	return cmd
}
