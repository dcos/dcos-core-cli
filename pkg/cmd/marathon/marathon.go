package marathon

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/app"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/debug"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/deployment"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/group"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/leader"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/pod"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/python"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/task"
	"github.com/spf13/cobra"
)

// NewCommand creates the `dcos package` subcommand.
func NewCommand(ctx api.Context) *cobra.Command {
	var configSchema bool
	var info bool
	var version bool

	cmd := &cobra.Command{
		Use:   "marathon",
		Short: "Deploy and manage applications to DC/OS",
		RunE: func(cmd *cobra.Command, args []string) error {
			/*
				if len(args) == 0 {
					return cmd.Help()
				}
				fmt.Fprintln(ctx.ErrOut(), cmd.UsageString())
				return fmt.Errorf("unknown command %s", args[0])
			*/
			return python.InvokePythonCLI(ctx)
		},
	}

	cmd.Flags().BoolVar(&configSchema, "config-schema", false, "Show the configuration schema for the Marathon subcommand.")
	cmd.Flags().BoolVar(&info, "info", false, "Print a short description of this subcommand.")
	cmd.Flags().BoolVar(&version, "version", false, "Print version information")

	cmd.AddCommand(
		newCmdMarathonAbout(ctx),
		newCmdMarathonDelay(ctx),
		newCmdMarathonPing(ctx),
		newCmdMarathonPlugin(ctx),
		app.NewCmdMarathonApp(ctx),
		debug.NewCmdMarathonDebug(ctx),
		deployment.NewCmdMarathonDeployment(ctx),
		group.NewCmdMarathonGroup(ctx),
		leader.NewCmdMarathonLeader(ctx),
		pod.NewCmdMarathonPod(ctx),
		task.NewCmdMarathonTask(ctx),
	)

	return cmd
}
