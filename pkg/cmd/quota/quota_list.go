package quota

import (
	"encoding/json"
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/mesos"
	"github.com/spf13/cobra"
)

func newCmdQuotaList(ctx api.Context) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List quotas",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := mesosClient(ctx)
			if err != nil {
				return err
			}

			roles, err := c.Roles()
			if err != nil {
				return err
			}

			allRoles := roles.Roles
			quotas := []mesos.Quota{}

			for _, role := range allRoles {
				if role.Quota.Role != "" {
					quotas = append(quotas, role.Quota)
				}
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(quotas)
			}

			for _, quota := range quotas {
				fmt.Fprintln(ctx.Out(), quota.Role)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	return cmd
}
