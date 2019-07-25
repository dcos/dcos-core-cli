package quota

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/mesos"
	"github.com/spf13/cobra"
)

func newCmdQuotaUpdate(ctx api.Context) *cobra.Command {
	var force bool
	var cpu, disk, gpu, mem float64

	cmd := &cobra.Command{
		Use:   "update <group>",
		Short: "Update a quota",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mesosClient, err := mesos.NewClientWithContext(ctx)
			if err != nil {
				return err
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
