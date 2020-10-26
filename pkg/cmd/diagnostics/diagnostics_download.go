package diagnostics

import (
	"fmt"
	"os"

	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newDiagnosticsDownloadCommand(ctx api.Context) *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		// the <bundle-id> is used to indicate what the expected positional arg is in the help output
		Use:   "download <bundle-id>",
		Short: "Download diagnostics bundle",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := client(ctx)
			if err != nil {
				return err
			}

			var id string
			if len(args) == 0 {
				bundle, err := latestBundle(client)
				if err != nil {
					return err
				}
				id = bundle.ID
			} else {
				id = args[0]
			}

			if outputPath == "" {
				outputPath = fmt.Sprintf("%s.zip", id)
			}

			outFile, err := os.Create(outputPath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			return client.Download(id, outFile)
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Set output path (defaults to '<bundle-id>.zip')")
	return cmd
}
