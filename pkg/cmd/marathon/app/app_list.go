package app

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/marathon"
	"github.com/dustin/go-humanize"
	m "github.com/gambol99/go-marathon"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

const defaultTableValue = "---"

// `truncate` will break if this is less than 3
const maxTableValueLength = 30

var deploymentDisplay = map[string]string{
	"ResolveArtifacts":   "artifacts",
	"ScaleApplication":   "scale",
	"StartApplication":   "start",
	"StopApplication":    "stop",
	"RestartApplication": "restart",
	"ScalePod":           "scale",
	"StopPod":            "stop",
	"RestartPod":         "restart",
	"KillAllOldTasksOf":  "kill-tasks",
}

func newCmdMarathonAppList(ctx api.Context) *cobra.Command {
	var jsonOutput bool
	var quiet bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the installed applications.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := marathon.NewClient(ctx)
			if err != nil {
				return err
			}

			applications, err := client.Applications()
			if err != nil {
				return err
			}

			if quiet {
				for _, app := range applications {
					fmt.Fprintln(ctx.Out(), app.ID)
				}
				return nil
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")

				return enc.Encode(&applications)
			}

			deployments, err := client.Deployments()
			if err != nil {
				return err
			}

			deploymentMap := make(map[string]m.Deployment)
			for _, d := range deployments {
				deploymentMap[d.ID] = d
			}

			queue, err := client.Queue()
			if err != nil {
				return err
			}

			tableHeader := []string{"ID", "MEM", "CPUS", "TASKS", "HEALTH", "DEPLOYMENT", "WAITING", "CONTAINER", "CMD", "ROLE"}
			table := cli.NewTable(ctx.Out(), tableHeader)
			table.SetAlignment(tablewriter.ALIGN_LEFT)

			items := make([][]string, 0, len(applications))
			for _, app := range applications {
				item := []string{
					app.ID,
					humanize.Ftoa(*app.Mem),
					humanize.Ftoa(app.CPUs),
					formatTaskRunning(app),
					formatHealth(app),
					formatDeployments(app, deploymentMap),
					formatWaiting(app, queue),
					formatContainerType(app),
					formatCmd(app),
					formatRole(app),
				}
				items = append(items, item)
			}
			sort.Sort(appRows(items))
			table.AppendBulk(items)
			table.Render()

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print JSON-formatted data.")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Display IDs only.")

	return cmd
}

func truncate(s string) string {
	if len(s) < maxTableValueLength {
		return s
	}

	return s[:maxTableValueLength-3] + "..."
}

func formatCmd(app m.Application) string {
	if app.Cmd != nil && *app.Cmd != "" {
		return truncate(*app.Cmd)
	}
	if app.Args != nil && len(*app.Args) > 0 {
		return truncate(fmt.Sprintf("[%s]", strings.Join(*app.Args, ", ")))
	}
	return defaultTableValue
}

func formatContainerType(app m.Application) string {
	// TODO: everything to upper/lower case?
	if app.Container != nil && app.Container.Type != "" {
		return app.Container.Type
	}

	return "mesos"
}

func formatTaskRunning(app m.Application) string {
	if app.Instances != nil {
		return fmt.Sprintf("%d/%d", app.TasksRunning, *app.Instances)
	}
	return fmt.Sprintf("%d/0", app.TasksRunning)
}

func formatHealth(app m.Application) string {
	if app.HealthChecks == nil {
		return defaultTableValue
	}
	return fmt.Sprintf("%d/%d", app.TasksHealthy, app.TasksRunning)
}

func formatDeployments(app m.Application, deployments map[string]m.Deployment) string {
	actions := []string{}

	for _, d := range app.Deployments {
		if id, ok := d["id"]; ok {
			if deployment, ok := deployments[id]; ok {
				for _, action := range deployment.CurrentActions {
					if action.App == app.ID {
						actions = append(actions, deploymentDisplay[action.Action])
					}
				}
			}
		}
	}

	if len(actions) == 0 {
		return defaultTableValue
	} else if len(actions) == 1 {
		return actions[0]
	}
	return fmt.Sprintf("(%s)", strings.Join(actions, ","))
}

func formatWaiting(app m.Application, queue m.Queue) string {
	for _, item := range queue.Items {
		if app.ID == item.Application.ID {
			if item.Delay.Overdue {
				return "true"
			}
		}
	}
	return "false"
}

func formatRole(app m.Application) string {
	if app.Role == nil {
		return ""
	}
	return *app.Role
}

// this enables sorting by ID
type appRows [][]string

func (a appRows) Swap(i int, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a appRows) Less(i int, j int) bool {
	return strings.Compare(a[i][0], a[j][0]) < 0
}

func (a appRows) Len() int {
	return len(a)
}
