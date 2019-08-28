package diagnostics

import (
	"time"

	"github.com/dcos/dcos-cli/api"
	diagnostics "github.com/dcos/dcos-core-cli/pkg/diagnostics/v2"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

const pollInterval time.Duration = time.Second

func newDiagnosticsWaitCommand(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wait <bundle-id>",
		Short: "Wait until the given bundle is completed",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := pluginutil.HTTPClient("")
			client := diagnostics.NewClient(c)

			// this seemed to be the easiest way to manage using two different methods
			// of getting the right bundle depending on the arguments
			bundle, err := func() (*diagnostics.Bundle, error) {
				if len(args) == 0 {
					return latestBundle(client)
				}
				return client.Get(args[0])
			}()
			if err != nil {
				return err
			}

			if len(args) == 0 {
				ctx.Logger().Infof("Waiting for bundle %s", bundle.ID)
			}

			for {
				if bundle.IsFinished() {
					break
				}

				time.Sleep(pollInterval)

				bundle, err = client.Get(bundle.ID)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	return cmd
}
