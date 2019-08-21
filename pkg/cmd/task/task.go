package task

import (
	"fmt"
	"strings"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/mesos"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/docker/docker/pkg/term"
	"github.com/gobwas/glob"
	mesosgo "github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli/httpagent"
	"github.com/spf13/cobra"
)

// NewCommand creates the `core service` subcommand.
func NewCommand(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task <task-id>",
		Short: "Manage DC/OS tasks",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, ok := ctx.EnvLookup(cli.EnvStrictDeprecations)
			if !ok {
				ctx.Deprecated("Getting the list of tasks from `dcos task` is deprecated. Please use `dcos task list`.")
				listCmd := newCmdTaskList(ctx)
				// Execute by default would use os.Args[1:], which is everything after `dcos ...`.
				// We need all command line arguments after `dcos task ...`.
				listCmd.SetArgs(ctx.Args()[2:])
				listCmd.SilenceErrors = true
				listCmd.SilenceUsage = true
				return listCmd.Execute()
			}
			return cmd.Help()
		},
	}
	cmd.Flags().Bool("all", false, "Print completed and in-progress tasks")
	cmd.Flags().Bool("json", false, "Print in json format")
	cmd.Flags().Bool("completed", false, "Print completed tasks")

	cmd.AddCommand(
		newCmdTaskAttach(ctx),
		newCmdTaskDownload(ctx),
		newCmdTaskExec(ctx),
		newCmdTaskList(ctx),
		newCmdTaskLog(ctx),
		newCmdTaskLs(ctx),
		newCmdTaskMetrics(ctx),
	)
	return cmd
}

type taskFilters struct {
	Active    bool
	Completed bool
	ID        string
	Agent     string
}

func findTask(ctx api.Context, filters taskFilters) (*mesos.Task, error) {
	tasks, err := findTasks(ctx, filters)
	if err != nil {
		return nil, err
	}

	if len(tasks) > 1 {
		var names []string
		for _, task := range tasks {
			if task.ID == filters.ID {
				return &task, nil
			}
			names = append(names, task.ID)
		}
		return nil, fmt.Errorf("found more than one task with the same name: %v", names)
	}

	return &tasks[0], nil
}

func findTasks(ctx api.Context, filters taskFilters) ([]mesos.Task, error) {
	mesosClient, err := mesos.NewClientWithContext(ctx)
	if err != nil {
		return nil, err
	}

	state, err := mesosClient.State()
	if err != nil {
		return nil, err
	}

	var g glob.Glob
	if filters.ID != "" {
		g, err = glob.Compile(filters.ID)
		if err != nil {
			return nil, err
		}
	}

	tasks := []mesos.Task{}
	for _, f := range state.Frameworks {
		for _, t := range f.Tasks {
			if filters.Active && matchTask(t, filters, g) {
				tasks = append(tasks, t)
			}
		}
		for _, t := range f.CompletedTasks {
			if filters.Completed && matchTask(t, filters, g) {
				tasks = append(tasks, t)
			}
		}
	}

	if len(tasks) == 0 {
		if filters.ID != "" && filters.Agent != "" {
			return tasks, fmt.Errorf("no task ID found containing '%s' in agent '%s'", filters.ID, filters.Agent)
		}
		if filters.ID != "" {
			return tasks, fmt.Errorf("no task ID found containing '%s'", filters.ID)
		}
		if filters.Agent != "" {
			return tasks, fmt.Errorf("no task found in agent '%s'", filters.Agent)
		}
	}
	return tasks, nil
}

func matchTask(task mesos.Task, filters taskFilters, g glob.Glob) bool {
	if filters.Agent != "" && task.SlaveID != filters.Agent {
		return false
	}
	if filters.ID == "" {
		return true
	}
	return strings.Contains(task.ID, filters.ID) || g.Match(task.ID)
}

func mesosHTTPClient(ctx api.Context, agentID string) (*httpcli.Client, error) {
	cluster, err := ctx.Cluster()
	if err != nil {
		return nil, err
	}

	mesosURL, _ := cluster.Config().Get("core.mesos_master_url").(string)
	if mesosURL != "" {
		mesosClient, err := mesos.NewClientWithContext(ctx)
		if err != nil {
			return nil, err
		}

		agents, err := mesosClient.Agents()
		if err != nil {
			return nil, err
		}

		for _, a := range agents {
			if a.AgentInfo.ID.Value == agentID {

				// PID format: Host@IP:Port
				// TODO: AgentInfo provides a hostname and a port, investigate if they can be used instead.
				ipPort := strings.Split(a.GetPID(), "@")
				url := fmt.Sprintf("http://%s", ipPort)

				return httpcli.New(httpcli.Endpoint(url)), nil
			}
		}
		return nil, fmt.Errorf("Agent ID %s not found", agentID)
	}

	rt := pluginutil.NewHTTPClient().Transport

	httpClient := httpcli.New(
		httpcli.Endpoint(fmt.Sprintf("%s/slave/%s/api/v1", cluster.URL(), agentID)),
		httpcli.Do(httpcli.With(httpcli.RoundTripper(rt))),
	)
	return httpClient, nil
}

func newTaskIO(ctx api.Context, id string, interactive bool, tty bool, user string) (*mesos.TaskIO, error) {
	filters := taskFilters{
		Active: true,
		ID:     id,
	}

	task, err := findTask(ctx, filters)
	if err != nil {
		return nil, err
	}

	httpClient, err := mesosHTTPClient(ctx, task.SlaveID)
	if err != nil {
		return nil, err
	}

	containerID := mesosgo.ContainerID{
		Value: task.Statuses[0].ContainerStatus.ContainerID.Value,
	}

	if task.Statuses[0].ContainerStatus.ContainerID.Parent != nil {
		containerID.Parent = &mesosgo.ContainerID{
			Value: task.Statuses[0].ContainerStatus.ContainerID.Parent.Value,
		}
	}

	opts := mesos.TaskIOOpts{
		Stdin:       ctx.Input(),
		Stdout:      ctx.Out(),
		Stderr:      ctx.ErrOut(),
		Interactive: interactive,
		TTY:         tty,
		User:        user,
		Sender:      httpagent.NewSender(httpClient.Send),
		Logger:      pluginutil.Logger(),
	}

	if escapeSequenceEnv, ok := ctx.EnvLookup("DCOS_TASK_ESCAPE_SEQUENCE"); ok {
		opts.EscapeSequence, err = term.ToBytes(escapeSequenceEnv)
		if err != nil {
			return nil, err
		}
	}
	return mesos.NewTaskIO(containerID, opts)
}
