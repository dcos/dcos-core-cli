package task

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

// newCmdTaskMetrics creates the `task metrics` command with all its subcommands.
func newCmdTaskMetrics(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Print the metrics of a task",
	}

	cmd.AddCommand(
		newCmdTaskMetricsDetails(ctx),
		newCmdTaskMetricsSummary(ctx),
	)

	return cmd
}
