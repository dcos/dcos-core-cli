package node

import (
	"encoding/json"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/spf13/cobra"
)

func newCmdNodeDNS(ctx api.Context) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "dns <hostname>",
		Short: "Return the IP address(es) corresponding to a given hostname",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hosts, err := mesosClient().Hosts(args[0])
			if err != nil {
				return err
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(hosts)
			}

			table := cli.NewTable(ctx.Out(), []string{"HOST", "IP"})
			for _, host := range hosts {
				table.Append([]string{host.Host, host.IP})
			}
			table.Render()

			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	return cmd
}
