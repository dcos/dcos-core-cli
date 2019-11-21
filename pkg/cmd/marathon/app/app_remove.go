package app

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/marathon"
	goMarathon "github.com/gambol99/go-marathon"
	"github.com/spf13/cobra"
)

func newCmdMarathonAppRemove(ctx api.Context) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove an application.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := marathon.NewClient(ctx)
			if err != nil {
				return err
			}

			return marathonAppRemove(*client, args[0], force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Disable checks in Marathon during updates.")

	return cmd
}

func marathonAppRemove(client marathon.Client, appID string, force bool) error {
	_, err := client.API.DeleteApplication(appID, force)
	if apiErr, ok := err.(*goMarathon.APIError); ok && apiErr.ErrCode == goMarathon.ErrCodeNotFound {
		return fmt.Errorf(`app '/%s' does not exist`, appID)
	}
	return err
}
