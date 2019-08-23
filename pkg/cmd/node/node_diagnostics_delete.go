package node

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/spf13/cobra"
)

func newCmdNodeDiagnosticsDelete(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <bundle>",
		Short: "Delete a diagnostics bundle",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, ok := ctx.EnvLookup(cli.EnvStrictDeprecations)
			if ok {
				return fmt.Errorf("unknown command delete")
			}
			ctx.Deprecated("This command is deprecated since DC/OS 1.14, please use 'dcos diagnostics delete' instead.")

			resp, err := diagnosticsClient().Delete(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", resp.Status)
			return nil
		},
	}
	return cmd
}
