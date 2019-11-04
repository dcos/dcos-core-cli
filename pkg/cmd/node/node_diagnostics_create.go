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
			err := ctx.Deprecated("This command is deprecated since DC/OS 2.0, please use 'dcos diagnostics create' instead.")
			if err != nil {
				return err
			}

			resp, err := diagnosticsClient().Create(args)
			if err != nil {
				return err
			}
			fmt.Printf("%s, available bundle: %s\n", resp.Status, resp.Extra.BundleName)
			return nil
		},
	}
	return cmd
}
