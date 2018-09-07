package job

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

// newCmdClusterRun runs a given job right now.
func newCmdJobAdd(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "add a job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := metronomeClient(ctx)
			if err != nil {
				return err
			}

			_, err = client.Add(args[0])
			return err
		},
	}
	return cmd
}
