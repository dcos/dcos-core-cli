package marathon

import (
	"encoding/json"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/marathon"

	"github.com/spf13/cobra"
)

func newCmdMarathonAbout(ctx api.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "about",
		Short: "Print info.json for DC/OS Marathon.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := marathon.NewClient(ctx)
			if err != nil {
				return err
			}

			return marathonAbout(ctx, *client)
		},
	}
}

func marathonAbout(ctx api.Context, client marathon.Client) error {
	info, err := client.Info()
	if err != nil {
		return err
	}

	enc := json.NewEncoder(ctx.Out())
	enc.SetIndent("", "    ")
	return enc.Encode(info)
}
