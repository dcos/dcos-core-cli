package node

import (
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/spf13/cobra"
)

func newCmdNodeDiagnosticsList(ctx api.Context) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available diagnostics bundles",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, ok := ctx.EnvLookup(cli.EnvStrictDeprecations)
			if ok {
				return fmt.Errorf("unknown command list")
			}
			ctx.Deprecated("This command is deprecated since DC/OS 1.14, please use 'dcos diagnostics list' instead.")

			list, err := diagnosticsClient().List()
			if err != nil {
				return err
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(list)
			}

			fmt.Println("Available diagnostic bundles:")
			for _, bundles := range list {
				for _, bundle := range bundles {
					fmt.Printf("%s %.1fMiB\n", filepath.Base(bundle.File), float64(bundle.Size)/math.Pow(10, 6))
				}

			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	return cmd
}
