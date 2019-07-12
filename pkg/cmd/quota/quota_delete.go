package quota

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdQuotaDelete(ctx api.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "create <group>",
		Short: "Delete a quota",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet")
		},
	}
}
