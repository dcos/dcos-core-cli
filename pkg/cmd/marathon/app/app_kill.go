package app

import (
	"encoding/json"
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/marathon"
	goMarathon "github.com/gambol99/go-marathon"
	"github.com/spf13/cobra"
)

func newCmdMarathonAppKill(ctx api.Context) *cobra.Command {
	var scale bool
	var host string

	cmd := &cobra.Command{
		Use:   "kill",
		Short: "Kill a running application instance.",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if scale && host != "" {
				return fmt.Errorf("the flags 'scale' and 'host' cannot be used at the same time")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := marathon.NewClient(ctx)
			if err != nil {
				return err
			}

			return marathonAppKill(ctx, *client, args[0], scale, host)
		},
	}

	cmd.Flags().BoolVar(&scale, "scale", false, "Scale the app down after performing the operation.")
	cmd.Flags().StringVar(&host, "host", "", "The hostname that is running app.")

	return cmd
}

func marathonAppKill(ctx api.Context, client marathon.Client, appID string, scale bool, host string) error {
	if scale {
		_, err := client.API.Application(appID)
		if apiErr, ok := err.(*goMarathon.APIError); ok && apiErr.ErrCode == goMarathon.ErrCodeNotFound {
			return fmt.Errorf(`path '/%s' does not exist`, appID)
		}
		if err != nil {
			return err
		}

		result, err := client.API.ScaleApplicationInstances(appID, 0, false)
		if err != nil {
			return err
		}

		deployment, err := json.Marshal(result)
		if err != nil {
			return err
		}

		fmt.Fprintf(ctx.Out(), "Started deployment: %s\n", string(deployment))
		return nil
	}

	// go-marathon KillApplicationTasks() does not return
	// enough information thus we do not use it.
	result, err := client.KillTasks(appID, host)
	if err != nil {
		return err
	}

	killedTasks, err := json.Marshal(result["tasks"].([]interface{}))
	if err != nil {
		return err
	}

	fmt.Fprintf(ctx.Out(), "Killed tasks: %s\n", string(killedTasks))
	return nil
}
