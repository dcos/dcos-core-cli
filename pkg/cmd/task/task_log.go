package task

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/logs"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

func newCmdTaskLog(ctx api.Context) *cobra.Command {
	// all and completed are useless but we keep them for retrocompatibility.
	var all, completed, follow bool
	var lines int
	var file, output string

	cmd := &cobra.Command{
		Use:   "log <task> [<file>]",
		Short: "Print the task log. By default, the 10 most recent task logs from stdout are printed.",
		Args:  cobra.RangeArgs(1, 2),
		PreRun: func(cmd *cobra.Command, args []string) {
			switch len(args) {
			case 1:
				file = "stdout"
			case 2:
				file = args[1]
			}

			// We support negative lines for backwards compatibility with the Python CLI.
			// The underlying API takes negative numbers so some users may rely on this.
			if lines < 0 {
				lines *= -1
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			filters := taskFilters{
				Active:    !completed,
				Completed: all || completed,
				ID:        args[0],
			}

			tasks, err := findTasks(ctx, filters)
			if err != nil {
				return err
			}

			// Only one task matching or multiple tasks but no follow.
			if len(tasks) == 1 || follow == false {
				if output != "short" {
					output = "short"
					ctx.Logger().Info(`Task logs don't support output options. Defaulting to "short"...`)
				}

				var failed bool
				for _, task := range tasks {
					if len(tasks) > 1 {
						fmt.Fprintln(ctx.Out(), fmt.Sprintf("===> %s <===", task.ID))
					}
					logClient := logs.NewClient(pluginutil.HTTPClient(""), ctx.Out())
					opts := logs.Options{
						Follow: follow,
						Format: output,
						Skip:   -1 * lines,
					}
					err := logClient.PrintTask(task.ID, file, opts)
					if err != nil {
						failed = true
						fmt.Fprintf(ctx.ErrOut(), "Error: %v\n", err)
					}
				}

				if len(tasks) > 1 && failed {
					return fmt.Errorf("could not log all matched tasks")
				}
				return nil
			}

			// Follow multiple tasks.
			if output != "cat" {
				ctx.Logger().Info(`Task logs don't support output options. Defaulting to "cat"...`)
			}

			// The channel receiving the content of the logs. Each task followed dumps its logs in it.
			msgChan := make(chan taskData)
			errChan := make(chan error)
			for _, task := range tasks {
				go func(taskID string, lines int, file string, c chan taskData, e chan error) {
					taskOut := &taskWriter{
						task:   taskID,
						writer: c,
					}
					logClient := logs.NewClient(pluginutil.HTTPClient(""), taskOut)
					opts := logs.Options{
						Follow: true,
						Format: "cat",
						Skip:   -1 * lines,
					}
					err := logClient.FollowTask(taskID, file, false, opts)
					if err != nil {
						e <- err
					}
				}(task.ID, lines, file, msgChan, errChan)
			}

			lastTask := ""
			for {
				select {
				case newTaskMsg := <-msgChan:
					// The new log line is coming from a task that is not the last one, print its ID.
					if newTaskMsg.task != lastTask {
						fmt.Fprintln(ctx.Out(), fmt.Sprintf("===> %s <===", newTaskMsg.task))
						lastTask = newTaskMsg.task
					}
					fmt.Fprintln(ctx.Out(), string(newTaskMsg.data))
				case err := <-errChan:
					return err
				}
			}
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "Print completed and in-progress tasks")
	cmd.Flags().BoolVar(&completed, "completed", false, "Print completed tasks")
	cmd.Flags().BoolVar(&follow, "follow", false, "Dynamically update the log")
	cmd.Flags().IntVar(&lines, "lines", 10, "Print the N last lines. 10 is the default")
	cmd.Flags().StringVarP(&output, "output", "o", "short", "Format log message output")
	return cmd
}

type taskData struct {
	task string
	data []byte
}

type taskWriter struct {
	task   string
	writer chan taskData
}

func (t *taskWriter) Write(data []byte) (n int, err error) {
	t.writer <- taskData{t.task, data}
	return len(data), nil
}
