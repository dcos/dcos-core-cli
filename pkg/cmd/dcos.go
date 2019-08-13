package cmd

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cmd/diagnostics"
	"github.com/dcos/dcos-core-cli/pkg/cmd/job"
	"github.com/dcos/dcos-core-cli/pkg/cmd/node"
	"github.com/dcos/dcos-core-cli/pkg/cmd/pkg"
	"github.com/dcos/dcos-core-cli/pkg/cmd/quota"
	"github.com/dcos/dcos-core-cli/pkg/cmd/service"
	"github.com/dcos/dcos-core-cli/pkg/cmd/task"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

const annotationUsageOptions string = "usage_options"

// NewDCOSCommand creates the `dcos` command with all the available subcommands.
func NewDCOSCommand(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use: "dcos",
	}

	cmd.AddCommand(
		job.NewCommand(ctx),
		node.NewCommand(ctx),
		pkg.NewCommand(ctx),
		quota.NewCommand(ctx),
		service.NewCommand(ctx),
		task.NewCommand(ctx),
		diagnostics.NewCommand(ctx),
	)

	cmd.SetUsageFunc(pluginutil.Usage)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	return cmd
}
