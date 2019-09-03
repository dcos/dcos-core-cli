package node

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdNodeDiagnosticsDownload(ctx api.Context) *cobra.Command {
	var location string
	cmd := &cobra.Command{
		Use:   "download <bundle>",
		Short: "Download a diagnostics bundle",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := ctx.Deprecated("This command is deprecated since DC/OS 1.14, please use 'dcos diagnostics download' instead.")
			if err != nil {
				return err
			}

			if location == "" {
				var e error
				location, e = os.Getwd()
				if e != nil {
					return err
				}
			}

			client := diagnosticsClient()

			isBundleFound := false
			list, err := client.List()
			if err != nil {
				return err
			}
			for _, bundles := range list {
				for _, bundle := range bundles {
					if filepath.Base(bundle.File) == args[0] {
						isBundleFound = true
					}
				}
			}
			if !isBundleFound {
				return fmt.Errorf("Unable to find bundle '%s'", args[0])
			}

			out, err := os.Create(filepath.Join(location, args[0]))
			if err != nil {
				return err
			}
			defer out.Close()

			resp, err := client.Get(args[0])
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			_, err = io.Copy(out, resp.Body)
			return err
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Where to download the diagnostics bundle")
	return cmd
}
