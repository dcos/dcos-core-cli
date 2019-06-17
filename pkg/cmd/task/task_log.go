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
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			tasks, err := findTasks(ctx, args[0])
			if err != nil {
				return err
			}

			// Only one task matching or multiple tasks but no follow.
			if len(tasks) == 1 || follow == false {
				for _, task := range tasks {
					if len(tasks) > 1 {
						fmt.Fprintln(ctx.Out(), fmt.Sprintf("===> %s <===", task.TaskID.Value))
					}
					logClient := logs.NewClient(pluginutil.HTTPClient(""), ctx.Out())
					if output != "short" {
						output = "short"
						ctx.Logger().Info(`Task logs don't support output options. Defaulting to "short"...`)
					}
					opts := logs.Options{
						Follow: follow,
						Format: output,
						Skip:   -1 * lines,
					}
					return logClient.PrintTask(task.TaskID.Value, file, opts)
				}
			}

			// TODO (DCOS_OSS-5153): multiple followed tasks.
			var taskNames []string
			for _, task := range tasks {
				taskNames = append(taskNames, task.TaskID.Value)
			}
			return fmt.Errorf("found more than one task with the same name, unable to follow them all: %v", taskNames)
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "Print completed and in-progress tasks")
	cmd.Flags().BoolVar(&completed, "completed", false, "Print completed tasks")
	cmd.Flags().BoolVar(&follow, "follow", false, "Dynamically update the log")
	cmd.Flags().IntVar(&lines, "lines", 10, "Print the N last lines. 10 is the default")
	cmd.Flags().StringVarP(&output, "output", "o", "short", "Format log message output")
	return cmd
}
