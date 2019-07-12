package quota

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdQuotaCreate(ctx api.Context) *cobra.Command {
	var force bool
	var cpu, mem float32

	cmd := &cobra.Command{
		Use:   "create <group>",
		Short: "Create a quota",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet")
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force the quota creation")
	cmd.Flags().Float32Var(&cpu, "cpu", 1.0, "Amount of CPU for the quota")
	cmd.Flags().Float32Var(&mem, "mem", 1.0, "Amount of memory for the quota")
	return cmd
}
