package diagnostics

import (
	"encoding/json"
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	diagnostics "github.com/dcos/dcos-core-cli/pkg/diagnostics/v2"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

func newDiagnosticsListCommand(ctx api.Context) *cobra.Command {
	var jsonOutput bool
	var quietOutput bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available bundles",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := pluginutil.HTTPClient("")
			client := diagnostics.NewClient(c)
			bundles, err := client.List()
			if err != nil {
				return err
			}

			if quietOutput {
				for _, b := range bundles {
					if b.Type == diagnostics.Cluster {
						fmt.Fprintln(ctx.Out(), b.ID)
					}
				}
				return nil
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(bundles)
			}

			tableHeader := []string{"ID", "STATUS", "CREATED", "SIZE"}
			table := cli.NewTable(ctx.Out(), tableHeader)

			for _, b := range bundles {
				// skip any local bundles
				if b.Type == diagnostics.Local {
					continue
				}

				size := humanize.Bytes(uint64(b.Size))
				created := humanize.Time(b.Started)

				fields := []string{b.ID, b.Status.String(), created, size}
				table.Append(fields)
			}

			table.Render()

			return nil
		},
	}

	cmd.Flags().BoolVarP(&quietOutput, "quiet", "q", false, "Print only bundle IDs")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")

	return cmd
}
