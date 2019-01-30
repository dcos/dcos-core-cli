package node

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

// newCmdNodeMetrics creates the `core node metrics` subcommand.
func newCmdNodeMetrics(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Display the metrics of a node",
	}

	cmd.AddCommand(
		newCmdNodeMetricsDetails(ctx),
		newCmdNodeMetricsSummary(ctx),
	)
	return cmd
}
