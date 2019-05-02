package service

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/mesos"
	"github.com/spf13/cobra"
)

func newCmdServiceShutdown(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shutdown <service-name>",
		Short: "Shutdown a service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mesos.NewClientWithContext(ctx)
			if err != nil {
				return err
			}

			return client.TeardownFramework(args[0])
		},
	}
	return cmd
}
