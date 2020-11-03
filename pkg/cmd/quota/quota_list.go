package quota

import (
	"encoding/json"
	"fmt"
	"io"

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

			return listQuota(marathonClient, mesosClient, jsonOutput, ctx.Out())
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	return cmd
}

func listQuota(marathonClient *marathon.Client, mesosClient *mesos.Client, jsonOutput bool, out io.Writer) error {
	groupsResultChan := make(chan groupsResult)
	go func() {
		groups, err := marathonClient.GroupsWithoutRootSlash()
		groupsResultChan <- groupsResult{groups, err}
	}()

	rolesResChan := make(chan rolesResult)
	go func() {
		roles, err := mesosClient.Roles()
		rolesResChan <- rolesResult{roles, err}
	}()

	rolesRes := <-rolesResChan
	if rolesRes.err != nil {
		return rolesRes.err
	}

	groupsRes := <-groupsResultChan
	if groupsRes.err != nil {
		return groupsRes.err
	}

	var quotas []mesos.Quota
	for _, role := range rolesRes.roles.Roles {
		if role.Quota.Role != "" && groupsRes.groups[role.Quota.Role] {
			quotas = append(quotas, role.Quota)
		}
	}

	if jsonOutput {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "    ")
		return enc.Encode(quotas)
	}

	tableHeader := []string{"NAME", "CPU CONSUMED", "MEMORY CONSUMED", "DISK CONSUMED", "GPU CONSUMED"}
	table := cli.NewTable(out, tableHeader)

	for _, quota := range quotas {
		quotaInfo := map[string]string{"cpus": "", "mem": "", "disk": "", "gpus": ""}

		for info := range quotaInfo {
			limit, ok := quota.Limit[info].(float64)
			if !ok {
				quotaInfo[info] = "No limit"
				continue
			}
			var cons float64
			if consumed, ok := quota.Consumed[info].(float64); ok {
				cons = consumed
			}
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

			percent := ""
			if limit != 0 {
				percent = fmt.Sprintf("%.2f%% ", cons*100/limit)
			}
			quotaInfo[info] = fmt.Sprintf("%s(%s)", percent, desc)
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
}
