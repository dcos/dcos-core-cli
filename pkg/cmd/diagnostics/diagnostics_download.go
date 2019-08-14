package diagnostics

import (
	"fmt"
	"os"

	"github.com/dcos/dcos-cli/api"
	diagnostics "github.com/dcos/dcos-core-cli/pkg/diagnostics/v2"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

func newDiagnosticsDownloadCommand(ctx api.Context) *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		// the <bundle-id> is used to indicate what the expected positional arg is in the help output
		Use:   "download <bundle-id>",
		Short: "Download diagnostics bundle",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := pluginutil.HTTPClient("")
			client := diagnostics.NewClient(c)

			id := args[0]

			if outputPath == "" {
				outputPath = fmt.Sprintf("%s.zip", id)
			}

			outFile, err := os.Create(outputPath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			err = client.Download(id, outFile)
			if err != nil {
				outFile.Close()
				os.Remove(outputPath)

				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Set output path (defaults to bundle's ID in current directory)")
	return cmd
}
