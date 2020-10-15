package diagnostics

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	diagnostics "github.com/dcos/dcos-core-cli/pkg/diagnostics/v2"
	"github.com/spf13/cobra"
)

func newDiagnosticsCreateCommand(ctx api.Context) *cobra.Command {
	var agents, masters bool
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a diagnostics bundle",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := client(ctx)
			if err != nil {
				return err
			}
			opts := diagnostics.Options{
				Masters: masters || !masters && !agents,
				Agents:  agents || !masters && !agents,
			}
			id, err := client.Create(opts)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(ctx.Out(), id)
			return err
		},
	}
	cmd.Flags().BoolVar(&masters, "masters", false, "Collect data from masters")
	cmd.Flags().BoolVar(&agents, "agents", false, "Collect data from agents")

	return cmd
}
