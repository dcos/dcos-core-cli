package quota

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/marathon"
	"github.com/dcos/dcos-core-cli/pkg/mesos"
	"github.com/spf13/cobra"
)

func newCmdQuotaCreate(ctx api.Context) *cobra.Command {
	var force bool
	var cpu, disk, gpu, mem float64

	cmd := &cobra.Command{
		Use:   "create <group>",
		Short: "Create a quota",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if strings.Contains(args[0], "/") {
				return fmt.Errorf("could not create quota '%s', quotas cannot be created for subgroups yet", args[0])
			}

			if cpu == -1.0 && disk == -1.0 && gpu == -1.0 && mem == -1.0 {
				return fmt.Errorf("could not create quota '%s', at least one limit has to be given", args[0])
			}
			return nil
		},
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

			rResult := <-rolesRes
			if rResult.err != nil {
				return rResult.err
			}

			gResult := <-groupsRes
			if gResult.err != nil {
				return gResult.err
			}

			if !gResult.groups[args[0]] {
				return errors.New("/" + args[0] + " is not an existing Marathon group")
			}

			for _, role := range rResult.roles.Roles {
				// Deleted quotas can still exist but will have no limits set thus we check the content of the quota.
				if role.Quota.Role == args[0] && len(role.Quota.Limit) > 0 {
					return errors.New(args[0] + " is an existing quota, use 'dcos quota update' if you want to update it.")
				}
			}

			return mesosClient.UpdateQuota(args[0], cpu, disk, gpu, mem, force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force the quota creation")
	cmd.Flags().Float64Var(&cpu, "cpu", -1.0, "Number of CPUs for the quota's limit")
	cmd.Flags().Float64Var(&disk, "disk", -1.0, "Amount of disk (in MiB) for the quota's limit")
	cmd.Flags().Float64Var(&gpu, "gpu", -1.0, "Number of GPUs for the quota's limit")
	cmd.Flags().Float64Var(&mem, "mem", -1.0, "Amount of memory (in MiB) for the quota's limit")
	return cmd
}
