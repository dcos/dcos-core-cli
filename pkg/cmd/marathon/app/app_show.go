package app

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/marathon"
	"github.com/spf13/cobra"
)

const appVersionDescription = `The version of the application to use. It can be specified as an
absolute or relative value. Absolute values must be in ISO8601 date
format. Relative values must be specified as a negative integer and they
represent the version from the currently deployed application definition.
`

func newCmdMarathonAppShow(ctx api.Context) *cobra.Command {
	var appVersion string

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show the `marathon.json` for an  application.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := marathon.NewClient(ctx)
			if err != nil {
				return err
			}
			appID := args[0]

			targetVersion := ""
			if appVersion != "" {
				targetVersion, err = calculateVersion(client, appID, appVersion)
				if err != nil {
					return err
				}
			}

			app, err := client.ApplicationByVersion(appID, targetVersion)
			if err != nil {
				return err
			}

			enc := json.NewEncoder(ctx.Out())
			enc.SetIndent("", "    ")
			return enc.Encode(app)
		},
	}

	cmd.Flags().StringVar(&appVersion, "app-version", "", appVersionDescription)

	return cmd
}

func calculateVersion(client *marathon.Client, appID string, version string) (string, error) {
	// check if the version string is an integer which indicates that the user
	// wants that many behind the latest version
	versionsBehind, err := strconv.Atoi(version)
	if err != nil {
		// if it fails to parse, the user must have given a specific version string
		return version, nil
	}

	if versionsBehind >= 0 {
		return "", fmt.Errorf("relative versions must be negative: %d", versionsBehind)
	}
	versionsBehind = -1 * versionsBehind

	versions, err := client.ApplicationVersions(appID)
	if err != nil {
		return "", err
	}
	if len(versions.Versions) <= versionsBehind {
		return "", fmt.Errorf("application %s only has %d versions", appID, len(versions.Versions))
	}
	return versions.Versions[versionsBehind], nil
}
