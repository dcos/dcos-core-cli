package quota

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/spf13/cobra"
)

type resource struct {
	consumed string
	limit    string
}

func newCmdQuotaGet(ctx api.Context) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
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
			quotaFound := false

			for _, role := range roles.Roles {
				if role.Quota.Role == args[0] {
					quotaFound = true

					if jsonOutput {
						enc := json.NewEncoder(ctx.Out())
						enc.SetIndent("", "    ")
						return enc.Encode(role.Quota)
					}

					for resName, resLimit := range role.Quota.Limit {
						resourceNames = append(resourceNames, resName)
						var res resource
						res.consumed = formatValue(role.Quota.Consumed[resName], resName)
						res.limit = formatValue(resLimit, resName)

						if res.consumed != "-" && res.limit != "-" {
							if consumed, ok := role.Quota.Consumed[resName].(float64); ok {
								if limit, ok := role.Quota.Limit[resName].(float64); ok {
									res.consumed = fmt.Sprintf("%.2f%%", consumed*100/limit)
								}
							}
						}
						resources[resName] = res
					}
				}
			}

			if !quotaFound {
				return fmt.Errorf("could not find quota '%s'", args[0])
			}

			// Not exposing the guarantees yet as it's not displayed in the DC/OS UI.
			tableHeader := []string{"RESOURCE", "CONSUMPTION", "LIMIT"}
			table := cli.NewTable(ctx.Out(), tableHeader)

			sort.Strings(resourceNames)
			for _, name := range resourceNames {
				tableItem := []string{
					name,
					resources[name].consumed,
					resources[name].limit,
				}
				table.Append(tableItem)
			}

			table.Render()
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	return cmd
}

func formatValue(val interface{}, name string) string {
	switch typedVal := val.(type) {
	case float64:
		if name == "mem" || name == "disk" {
			if typedVal < 1024 {
				return fmt.Sprintf("%.0f MiB", typedVal)
			}
			if typedVal == float64(int64(typedVal)) {
				return fmt.Sprintf("%.0f GiB", typedVal/1024)
			}
			return fmt.Sprintf("%.1f GiB", typedVal/1024)
		}
		return fmt.Sprintf("%.0f", typedVal)
	}
	return "-"
}
