package diagnostics

import (
	"encoding/json"
	"fmt"

	"github.com/dcos/dcos-cli/api"
	diagnostics "github.com/dcos/dcos-core-cli/pkg/diagnostics/v2"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

func newDiagnosticsCreateCommand(ctx api.Context) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a diagnostics bundle",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := pluginutil.HTTPClient("")
			client := diagnostics.NewClient(c)

			id, err := client.Create()
			if err != nil {
				return err
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(id)
			}

			_, err = fmt.Fprintf(ctx.Out(), "Job has been successfully started, available bundle: %s\n", id)
			return err
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print the new bundle ID in JSON format")

	return cmd
}
