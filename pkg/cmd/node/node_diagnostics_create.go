package node

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdNodeDiagnosticsCreate(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create (<node IPs>)...",
		Short: "Create a diagnostics bundle",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := diagnosticsClient().Create(args)
			if err != nil {
				return err
			}
			fmt.Fprintf(ctx.Out(), "%s, available bundle: %s\n", resp.Status, resp.Extra.BundleName)
			return nil
		},
	}
	return cmd
}
