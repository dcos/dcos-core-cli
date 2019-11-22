package app

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/marathon"
	goMarathon "github.com/gambol99/go-marathon"
	"github.com/spf13/cobra"
)

func newCmdMarathonAppStop(ctx api.Context) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "stop",
		Args:  cobra.ExactArgs(1),
		Short: "Stop an application.",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := args[0]

			return appStop(ctx, appID, force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Disable checks in Marathon during updates.")

	return cmd
}

func appStop(ctx api.Context, appID string, force bool) error {
	client, err := marathon.NewClient(ctx)
	if err != nil {
		return err
	}
	description, err := client.API.Application(appID)
	if err != nil {
		if apiErr, ok := err.(*goMarathon.APIError); ok && apiErr.ErrCode == goMarathon.ErrCodeNotFound {
			return fmt.Errorf(`app '%s' does not exist`, appID)
		}
		return err
	}

	fmt.Println(description.Instances)
	if *description.Instances <= 0 {
		return fmt.Errorf("application '%s' already stopped: %d instances", appID, *description.Instances)
	}

	deploymentID, err := client.API.ScaleApplicationInstances(appID, 0, force)
	if err != nil {
		return err
	}

	fmt.Fprintf(ctx.Out(), "Created deployment %s\n", deploymentID.DeploymentID)

	return nil
}
