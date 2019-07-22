package quota

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdQuotaUpdate(ctx api.Context) *cobra.Command {
	var force bool
	var cpu, mem float64

	cmd := &cobra.Command{
		Use:   "update <group>",
		Short: "Update a quota",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet")
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force the quota creation")
	cmd.Flags().Float64Var(&cpu, "cpu", 1.0, "Amount of CPU for the quota's guarantee")
	cmd.Flags().Float64Var(&mem, "mem", 1.0, "Amount of memory (in MB) for the quota's guarantee")
	return cmd
}
