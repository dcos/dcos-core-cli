package group

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/python"
	"github.com/spf13/cobra"
)

const groupVersionDescription = `The group version to use for the command. It can be specified as an
absolute or relative value. Absolute values must be in ISO8601 date
format. Relative values must be specified as a negative integer and they
represent the version from the currently deployed group definition.`

func newCmdMarathonGroupShow(ctx api.Context) *cobra.Command {
	var groupVersion string

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Print a detailed list of groups.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return python.InvokePythonCLI(ctx)
		},
	}

	cmd.Flags().StringVar(&groupVersion, "group-version", "", groupVersionDescription)

	return cmd
}
