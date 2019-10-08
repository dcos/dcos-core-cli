package diagnostics

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	diagnostics "github.com/dcos/dcos-core-cli/pkg/diagnostics/v2"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

func newDiagnosticsCreateCommand(ctx api.Context) *cobra.Command {
	var masters bool
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a diagnostics bundle",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := pluginutil.HTTPClient("")
			client := diagnostics.NewClient(c)
			opts := diagnostics.Options{
				Masters: true,
				Agents:  true,
			}
			if masters {
				opts = diagnostics.Options{
					Masters: true,
					Agents:  false,
				}
			}
			id, err := client.Create(opts)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(ctx.Out(), id)
			return err
		},
	}
	cmd.Flags().BoolVar(&masters, "masters", true, "Collect data from masters only")

	return cmd
}
