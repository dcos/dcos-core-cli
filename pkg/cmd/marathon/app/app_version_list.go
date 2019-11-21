package app

import (
	"encoding/json"
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/marathon"
	"github.com/spf13/cobra"
)

func newCmdMarathonAppVersion(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Manage Marathon app versions.",
	}

	cmd.AddCommand(
		newCmdMarathonAppVersionList(ctx),
	)

	return cmd
}

func newCmdMarathonAppVersionList(ctx api.Context) *cobra.Command {
	var maxCount int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the version history of an application.",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("max-count") && maxCount <= 0 {
				return fmt.Errorf("maximum count must be a positive number")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := marathon.NewClient(ctx)
			if err != nil {
				return err
			}

			versions, err := marathonAppVersionList(*client, args[0], maxCount)
			if err != nil {
				return err
			}

			// versions is a slice of strings that has to be printed like
			// [
			//   "val1",
			//   "val2"
			// ]
			// To do so, we use the json package to marshal the slice
			// and then print it as a string.
			jsonVersions, err := json.MarshalIndent(versions, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(jsonVersions))

			return nil
		},
	}

	cmd.Flags().IntVar(&maxCount, "max-count", 0, "Maximum number of entries to fetch and return.")

	return cmd
}

func marathonAppVersionList(client marathon.Client, appID string, maxCount int) ([]string, error) {
	appVersions, err := client.API.ApplicationVersions(appID)
	if err != nil {
		return nil, err
	}

	if maxCount > 0 && maxCount <= len(appVersions.Versions) {
		return appVersions.Versions[:maxCount], nil
	}
	return appVersions.Versions, nil
}
