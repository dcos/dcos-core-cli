package quota

import (
	"encoding/json"
	"fmt"

	"github.com/dcos/dcos-cli/api"
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
				groups, err := marathonClient.GroupsAsQuotas()
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

			for _, quota := range quotas {
				fmt.Fprintln(ctx.Out(), quota.Role)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	return cmd
}
