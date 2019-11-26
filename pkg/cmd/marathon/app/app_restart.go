package app

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/marathon"
	goMarathon "github.com/gambol99/go-marathon"
	"github.com/spf13/cobra"
)

func newCmdMarathonAppRestart(ctx api.Context) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart an application.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := marathon.NewClient(ctx)
			if err != nil {
				return err
			}

			deploymentID, err := marathonAppRestart(*client, args[0], force)
			if apiErr, ok := err.(*goMarathon.APIError); ok && apiErr.ErrCode == goMarathon.ErrCodeNotFound {
				return fmt.Errorf(`app '/%s' does not exist`, args[0])
			}
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

func marathonAppRestart(client marathon.Client, appID string, force bool) (*goMarathon.DeploymentID, error) {
	app, err := client.API.Application(appID)
	if apiErr, ok := err.(*goMarathon.APIError); ok && apiErr.ErrCode == goMarathon.ErrCodeNotFound {
		return nil, fmt.Errorf(`app '/%s' does not exist`, appID)
	}
	if err != nil {
		return nil, err
	}

	if app.Instances == nil {
		return nil, fmt.Errorf("unable to get number of instances for application '/%s'", appID)
	}
	if *app.Instances <= 0 {
		return nil, fmt.Errorf("unable to perform rolling restart of application '/%s' because it has no running tasks", appID)
	}

	return client.API.RestartApplication(appID, force)
}
