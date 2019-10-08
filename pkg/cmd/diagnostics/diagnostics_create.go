package diagnostics

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	diagnostics "github.com/dcos/dcos-core-cli/pkg/diagnostics/v2"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

func newDiagnosticsCreateCommand(ctx api.Context) *cobra.Command {
	opts := diagnostics.Options{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a diagnostics bundle",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := pluginutil.HTTPClient("")
			client := diagnostics.NewClient(c)

			id, err := client.Create(opts)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(ctx.Out(), id)
			return err
		},
	}
	cmd.Flags().BoolVar(&opts.Masters, "masters", true, "Enable bundle collection on masters")
	cmd.Flags().BoolVar(&opts.Agents, "agents", true, "Enable bundle collection on agents")

	return cmd
}
