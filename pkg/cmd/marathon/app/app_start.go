package app

import (
	"fmt"
	"strconv"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/marathon"
	goMarathon "github.com/gambol99/go-marathon"
	"github.com/spf13/cobra"
)

func newCmdMarathonAppStart(ctx api.Context) *cobra.Command {
	var force bool
	var instances int

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start an application.",
		Args:  cobra.RangeArgs(1, 2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 2 {
				inst, err := strconv.Atoi(args[1])
				if err != nil {
					return err
				}

				instances = inst
			} else {
				instances = 1
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := marathon.NewClient(ctx)
			if err != nil {
				return err
			}

			deploymentID, err := marathonAppStart(*client, args[0], instances, force)
			if err != nil {
				return err
			}

			fmt.Fprintf(ctx.Out(), "Created deployment %s\n", deploymentID.DeploymentID)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Disable checks in Marathon during updates.")

	return cmd
}

func marathonAppStart(client marathon.Client, appID string, instances int, force bool) (*goMarathon.DeploymentID, error) {
	if instances <= 0 {
		return nil, fmt.Errorf("the number of instances must be positive: %d", instances)
	}

	app, err := client.API.Application(appID)
	if apiErr, ok := err.(*goMarathon.APIError); ok && apiErr.ErrCode == goMarathon.ErrCodeNotFound {
		return nil, fmt.Errorf(`app '/%s' does not exist`, appID)
	}
	if err != nil {
		return nil, err
	}

	if app.Instances == nil {
		return nil, fmt.Errorf("unable to get number of instances for application '%s'", appID)
	}
	if *app.Instances > 0 {
		return nil, fmt.Errorf("application '%s' already started: %d instances", appID, *app.Instances)
	}

	return client.API.ScaleApplicationInstances(appID, instances, force)
}
