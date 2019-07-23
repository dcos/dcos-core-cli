package quota

import (
	"errors"
	"fmt"

	"github.com/dcos/dcos-cli/api"
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
			c, err := mesosClient(ctx)
			if err != nil {
				return err
			}

			quota, err := c.Quota()
			if err != nil {
				return err
			}
			for _, quotaInfo := range quota.Status.Infos {
				if quotaInfo.GetRole() == args[0] {
					return errors.New(args[0] + " is an existing quota, use 'dcos quota update' if you want to update it.")
				}

			}

			return fmt.Errorf("not implemented yet")
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force the quota creation")
	cmd.Flags().Float64Var(&cpu, "cpu", 0.0, "Number of CPUs for the quota's guarantee")
	cmd.Flags().Float64Var(&disk, "disk", 0.0, "Amount of disk (in MB) for the quota's guarantee")
	cmd.Flags().Float64Var(&gpu, "gpu", 0.0, "Number of GPUs for the quota's guarantee")
	cmd.Flags().Float64Var(&mem, "mem", 0.0, "Amount of memory (in MB) for the quota's guarantee")
	return cmd
}
