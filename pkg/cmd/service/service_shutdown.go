package service

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/mesos"
	"github.com/spf13/cobra"
)

func newCmdServiceShutdown(ctx api.Context) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "shutdown <service-name>",
		Short: "Shutdown a service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceName := args[0]
			if serviceName == "" {
				return fmt.Errorf("service name must not be empty")
			}
			if !yes {
				err := ctx.Prompt().Confirm(fmt.Sprintf("Do you really want to teardown %s with all its tasks? [yes/no] ",
					serviceName), "no")
				if err != nil {
					return err
				}
			}

			client, err := mesos.NewClientWithContext(ctx)
			if err != nil {
				return err
			}

			return client.TeardownFramework(serviceName)
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Disable interactive mode and assume “yes” is the answer to all prompts")

	return cmd
}
