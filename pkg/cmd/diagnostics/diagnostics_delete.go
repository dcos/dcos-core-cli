package diagnostics

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newDiagnosticsDeleteCommand(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <bundle-id>",
		Short: "Delete a diagnostics bundle",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := client(ctx)
			if err != nil {
				return err
			}
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
