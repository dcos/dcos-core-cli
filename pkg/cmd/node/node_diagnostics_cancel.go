package node

import (
	"encoding/json"
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/spf13/cobra"
)

func newCmdNodeDiagnosticsCancel(ctx api.Context) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "cancel",
		Short: "Cancel a running diagnostics job",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, ok := ctx.EnvLookup(cli.EnvStrictDeprecations)
			if ok {
				return fmt.Errorf("unknown command cancel")
			}
			ctx.Deprecated("This command is deprecated since DC/OS 1.14, please use 'dcos diagnostics' instead.")

			resp, err := diagnosticsClient().Cancel()
			if err != nil {
				return err
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(resp)
			}
			fmt.Println(resp.Status)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	return cmd
}
