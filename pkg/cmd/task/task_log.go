package task

import (
	"fmt"
	"strings"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/logs"
	"github.com/dcos/dcos-core-cli/pkg/mesos"
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
			mesosClient, err := mesos.NewClientWithContext(ctx)
			if err != nil {
				return err
			}

			allTasks, err := mesosClient.Tasks()
			if err != nil {
				return err
			}

			var tasks []string
			for _, t := range allTasks {
				if strings.Contains(t.TaskID.Value, args[0]) {
					tasks = append(tasks, t.TaskID.Value)
				}
			}

			if len(tasks) == 0 {
				return fmt.Errorf("no task ID found containing '%s'", args[0])
			}

			// Only one task matching or multiple tasks but no follow.
			if len(tasks) == 1 || follow == false {
				for _, task := range tasks {
					if len(tasks) > 1 {
						fmt.Fprintln(ctx.Out(), fmt.Sprintf("===> %s <===", task))
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
					err = logClient.PrintTask(task, file, opts)
					if err != nil {
						return err
					}
				}
			} else {
				// TODO (DCOS_OSS-5153): multiple followed tasks.
				return fmt.Errorf("found more than one task with the same name, unable to follow them all: %v", tasks)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "Print completed and in-progress tasks")
	cmd.Flags().BoolVar(&completed, "completed", false, "Print completed tasks")
	cmd.Flags().BoolVar(&follow, "follow", false, "Dynamically update the log")
	cmd.Flags().IntVar(&lines, "lines", 10, "Print the N last lines. 10 is the default")
	cmd.Flags().StringVarP(&output, "output", "o", "short", "Format log message output")
	return cmd
}
