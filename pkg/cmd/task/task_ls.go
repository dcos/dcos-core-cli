package task

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-cli/pkg/httpclient"
	"github.com/dcos/dcos-core-cli/pkg/mesos"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

func newCmdTaskLs(ctx api.Context) *cobra.Command {
	// all and completed are useless but we keep them for retrocompatibility.
	var all, completed, long bool
	var path string

	cmd := &cobra.Command{
		Use:   "ls <task> [<path>]",
		Short: "Print the list of files in the Mesos task sandbox",
		Args:  cobra.RangeArgs(1, 2),
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 2 {
				path = args[1]
			} else {
				path = "."
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

			var containerParentIDs []string
			var executorIDs []string
			for _, t := range tasks {
				if t.Name == "" {
					return fmt.Errorf("unable to find task '%s'", args[0])
				}

				containerID := t.Statuses[0].ContainerStatus.ContainerID
				if containerID.Parent == nil {
					containerParentIDs = append(containerParentIDs, containerID.Value)
				} else {
					containerParentIDs = append(containerParentIDs, containerID.Parent.Value)
				}

				if t.ExecutorID != "" {
					executorIDs = append(executorIDs, t.ExecutorID)
				} else {
					executorIDs = append(executorIDs, t.ID)
				}
			}

			var agentsPaths = map[string]map[string]string{}
			client := mesos.NewClient(pluginutil.HTTPClient(""))
			for _, t := range tasks {
				if _, ok := agentsPaths[t.SlaveID]; !ok {
					agentPaths, err := client.Debug(t.SlaveID)
					if err != nil {
						return err
					}
					agentsPaths[t.SlaveID] = agentPaths
				}
			}

			for i, t := range tasks {
				if len(tasks) > 1 {
					fmt.Fprintln(ctx.Out(), "===> "+t.ID+" <===")
				}
				taskPath := "/frameworks/" + t.FrameworkID + "/executors/" + executorIDs[i] + "/runs/" + containerParentIDs[i]
				for key, value := range agentsPaths[t.SlaveID] {
					if strings.HasSuffix(key, taskPath) {
						ls, err := client.Browse(t.SlaveID, value+"/"+path)
						if err != nil {
							if httpErr, ok := err.(*httpclient.HTTPError); ok {
								if httpErr.Response.StatusCode == 404 {
									return fmt.Errorf("cannot access '%s': no such file or directory", path)
								}
							}
							return err
						}

						if long {
							tableHeader := []string{"MODE", "LINKS", "UID", "GID", "SIZE", "DATE", "NAME"}
							table := cli.NewTable(ctx.Out(), tableHeader)
							for _, el := range ls {
								item := []string{
									el.Mode,
									strconv.FormatFloat(el.NLink, 'f', 0, 64),
									el.UID,
									el.GID,
									strconv.FormatFloat(el.Size, 'f', 0, 64),
									time.Unix(int64(el.MTime), 0).Format(time.RFC822),
									filepath.Base(el.Path),
								}
								table.Append(item)
							}
							table.Render()
						} else {
							var entries []string
							for _, entry := range ls {
								entries = append(entries, filepath.Base(entry.Path))
							}
							fmt.Println(strings.Join(entries, "  "))
						}
					}
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "Print completed and in-progress tasks")
	cmd.Flags().BoolVar(&completed, "completed", false, "Print completed tasks")
	cmd.Flags().BoolVar(&long, "long", false, "Print full Mesos sandbox file attributes")
	return cmd
}
