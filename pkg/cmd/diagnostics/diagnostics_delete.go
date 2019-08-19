package diagnostics

import (
	"fmt"
	"github.com/dcos/dcos-cli/api"
	diagnostics "github.com/dcos/dcos-core-cli/pkg/diagnostics/v2"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

func newDiagnosticsDeleteCommand(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <bundle-id>",
		Short: "Delete a diagnostics bundle",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := pluginutil.HTTPClient("")
			client := diagnostics.NewClient(c)
			id := args[0]

			if err := client.Delete(id); err != nil {
				return err
			}
			fmt.Fprintf(ctx.Out(), "Bundle %s deleted\n", id)
			return nil
		},
	}

	return cmd
}
