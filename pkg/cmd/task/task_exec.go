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

func newCmdTaskExec(ctx api.Context) *cobra.Command {
	// all and completed are useless but we keep them for retrocompatibility.
	var interactive, tty bool

	cmd := &cobra.Command{
		Use:   "exec <task> [--] <cmd> [<args>...]",
		Short: "Launch a process (<cmd>) inside of a container for a task (<task>).",
		Args:  cobra.MinimumNArgs(2),
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
			containerID := task.Statuses[len(task.Statuses)-1].ContainerStatus.ContainerID
			nestedContainer := lib.ContainerID{
				Value:  b.String(),
				Parent: containerID,
			}

			client := mesos.NewClient(pluginutil.HTTPClient(""))
			err = client.LaunchNestedContainer(task.AgentID.Value, nestedContainer)
			if err != nil {
				return err
			}

			fmt.Println("Hello")
			return client.TaskAttachExec(task.AgentID.Value, nestedContainer, args[1], args[1:], false, false)
		},
	}
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Attach a STDIN stream to the remote command for an interactive session")
	cmd.Flags().BoolVar(&tty, "tty", false, "Print completed tasks")
	return cmd
}
