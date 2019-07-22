package quota

import (
	"fmt"
	"sort"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/spf13/cobra"
)

type resource struct {
	consumed  string
	guarantee string
	limit     string
}

func newCmdQuotaGet(ctx api.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Get a quota",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := mesosClient(ctx)
			if err != nil {
				return err
			}

			resourceNames := []string{}

			roles, err := c.Roles()
			if err != nil {
				return err
			}

			resources := map[string]resource{}

			for _, role := range roles.Roles {
				if role.Quota.Role == args[0] {
					for resName, resLimit := range role.Quota.Limit {
						resourceNames = append(resourceNames, resName)
						var res resource
						res.consumed = formatValue(role.Quota.Consumed[resName], resName)
						res.guarantee = formatValue(role.Quota.Guarantee[resName], resName)
						res.limit = formatValue(resLimit, resName)
						resources[resName] = res
					}
				}
			}

			tableHeader := []string{"RESOURCE", "CONSUMPTION", "GUARANTEE", "LIMIT"}
			table := cli.NewTable(ctx.Out(), tableHeader)

			sort.Strings(resourceNames)
			for _, name := range resourceNames {
				tableItem := []string{
					name,
					resources[name].consumed,
					resources[name].guarantee,
					resources[name].limit,
				}
				table.Append(tableItem)
			}

			table.Render()
			return nil
		},
	}
}

func formatValue(val interface{}, name string) string {
	switch typedVal := val.(type) {
	case float64:
		if name == "mem" || name == "disk" {
			return fmt.Sprintf("%.1fGiB", typedVal/1024)
		}
		return fmt.Sprintf("%.0f", typedVal)
	}
	return "-"
}
