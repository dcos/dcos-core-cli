package quota

import (
	"errors"

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

			quotaRes := make(chan quotaResult)
			go func() {
				quota, err := mesosClient.Quota()
				quotaRes <- quotaResult{quota, err}
			}()

			quotaResult := <-quotaRes
			if quotaResult.err != nil {
				return quotaResult.err
			}

			groupsResult := <-groupsRes
			if groupsResult.err != nil {
				return groupsResult.err
			}

			if !groupsResult.groups[args[0]] {
				return errors.New("/" + args[0] + " is not an existing group")
			}

			for _, quotaInfo := range quotaResult.quota.Status.Infos {
				if quotaInfo.GetRole() == args[0] {
					return errors.New(args[0] + " is an existing quota, use 'dcos quota update' if you want to update it.")
				}
			}

			return mesosClient.UpdateQuota(args[0], cpu, disk, gpu, mem, force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force the quota creation")
	cmd.Flags().Float64Var(&cpu, "cpu", 0.0, "Number of CPUs for the quota's limit")
	cmd.Flags().Float64Var(&disk, "disk", 0.0, "Amount of disk (in MiB) for the quota's limit")
	cmd.Flags().Float64Var(&gpu, "gpu", 0.0, "Number of GPUs for the quota's limit")
	cmd.Flags().Float64Var(&mem, "mem", 0.0, "Amount of memory (in MiB) for the quota's limit")
	return cmd
}
