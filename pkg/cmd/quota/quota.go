package quota

import (
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

// NewCommand creates the `dcos quota` subcommand.
func NewCommand(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quota",
		Short: "Manage DC/OS quotas",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			fmt.Fprintln(ctx.ErrOut(), cmd.UsageString())
			return fmt.Errorf("unknown command %s", args[0])
		},
	}

	cmd.AddCommand(
		newCmdQuotaCreate(ctx),
		newCmdQuotaDelete(ctx),
		newCmdQuotaGet(ctx),
		newCmdQuotaUpdate(ctx),
	)

	return cmd
}
