package app

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/python"
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
			return python.InvokePythonCLI(ctx)
		},
	}

	cmd.Flags().StringVar(&appVersion, "app-version", "", appVersionDescription)

	return cmd
}
