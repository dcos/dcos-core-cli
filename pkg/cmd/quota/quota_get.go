package quota

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdQuotaGet(ctx api.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "get <group>",
		Short: "Get a quota",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet")
		},
	}
}
