package group

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cmd/marathon/python"
	"github.com/spf13/cobra"
)

func newCmdMarathonGroupList(ctx api.Context) *cobra.Command {
	var json bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Print the list of groups.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return python.InvokePythonCLI(ctx)
		},
	}

	cmd.Flags().BoolVar(&json, "json", false, "Print JSON-formatted data.")

	return cmd
}
