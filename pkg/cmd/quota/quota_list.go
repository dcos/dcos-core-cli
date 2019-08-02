package quota

import (
	"encoding/json"
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/marathon"
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
			marathonClient, err := marathon.NewClient(ctx)
			if err != nil {
				return err
			}

			mesosClient, err := mesos.NewClientWithContext(ctx)
			if err != nil {
				return err
			}

			groupsRes := make(chan groupsResult)
			go func() {
				groups, err := marathonClient.GroupsWithoutRootSlash()
				groupsRes <- groupsResult{groups, err}
			}()

			rolesRes := make(chan rolesResult)
			go func() {
				roles, err := mesosClient.Roles()
				rolesRes <- rolesResult{roles, err}
			}()

			rolesResult := <-rolesRes
			if rolesResult.err != nil {
				return rolesResult.err
			}

			groupsResult := <-groupsRes
			if groupsResult.err != nil {
				return groupsResult.err
			}

			quotas := []mesos.Quota{}

			for _, role := range rolesResult.roles.Roles {
				if role.Quota.Role != "" && groupsResult.groups[role.Quota.Role] {
					quotas = append(quotas, role.Quota)
				}
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(quotas)
			}

			tableHeader := []string{"NAME", "CPU CONSUMED", "MEMORY CONSUMED", "DISK CONSUMED", "GPU CONSUMED"}
			table := cli.NewTable(ctx.Out(), tableHeader)

			for _, quota := range quotas {
				quotaInfo := map[string]string{"cpus": "", "mem": "", "disk": "", "gpus": ""}

				for info := range quotaInfo {
					if limit, ok := quota.Limit[info].(float64); ok {
						var cons float64
						if consumed, ok := quota.Consumed[info].(float64); ok {
							cons = consumed
						}
						percent := fmt.Sprintf("%.2f%%", cons*100/limit)
						var desc string
						switch info {
						case "cpus", "gpus":
							desc = fmt.Sprintf("%.0f of %.0f Cores", cons, limit)
						case "mem", "disk":
							if limit < 1000 {
								desc = fmt.Sprintf("%.0f MiB of %.0f MiB", cons, limit)
							} else {
								cons /= 1000
								consString := fmt.Sprintf("%.1f", cons)
								if cons == float64(int64(cons)) {
									consString = fmt.Sprintf("%.0f", cons)
								}

								limit /= 1000
								limitString := fmt.Sprintf("%.1f", limit)
								if limit == float64(int64(limit)) {
									limitString = fmt.Sprintf("%.0f", limit)
								}

								desc = fmt.Sprintf("%s GiB of %s GiB", consString, limitString)
							}
						}
						quotaInfo[info] = fmt.Sprintf("%s (%s)", percent, desc)
					} else {
						quotaInfo[info] = "No limit"
					}
				}

				tableItem := []string{
					quota.Role,
					quotaInfo["cpus"],
					quotaInfo["mem"],
					quotaInfo["disk"],
					quotaInfo["gpus"],
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
