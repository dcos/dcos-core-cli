package quota

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdQuotaDelete(ctx api.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <group>",
		Short: "Delete a quota",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := mesosClient(ctx)
			if err != nil {
				return err
			}

			return c.DeleteQuota(args[0])
		},
	}
}
