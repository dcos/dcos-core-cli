package node

import (
	"encoding/json"
	"strings"

	"github.com/olekukonko/tablewriter"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/mesos"
	"github.com/dcos/dcos-core-cli/pkg/networking"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

type stateResult struct {
	state *mesos.State
	err   error
}

type ipsResult struct {
	ips map[string][]string
	err error
}

func newCmdNodeList(ctx api.Context) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Show all nodes in the cluster",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := mesosClient()

			// The following code allows us to do the API calls for the
			// state, the public IPs, and the masters concurrently which
			// results in a significant speedup of the command.
			stateRes := make(chan stateResult)
			go func() {
				state, err := client.State()
				stateRes <- stateResult{state, err}
			}()

			ipsRes := make(chan ipsResult)
			go func() {
				c := networking.NewClient(pluginutil.HTTPClient(""))
				nodes, err := c.Nodes()
				if err != nil {
					ipsRes <- ipsResult{nil, err}
				}

				ips := make(map[string][]string)
				for _, node := range nodes {
					ips[node.PrivateIP] = node.PublicIPs
				}
				ipsRes <- ipsResult{ips, nil}
			}()

			masters, err := client.Masters()
			if err != nil {
				return err
			}

			ipsResult := <-ipsRes
			if ipsResult.err != nil {
				return err
			}
			ips := ipsResult.ips

			stateResult := <-stateRes
			if stateResult.err != nil {
				return err
			}
			state := stateResult.state

			// In order to create a nodes json object that contains masters and agents
			// we need a slice of interface{} that is able to contain both node types.
			nodes := make([]interface{}, 0)
			table := cli.NewTable(ctx.Out(), []string{"HOSTNAME", "IP", "PUBLIC IPS", "ID", "TYPE", "REGION", "ZONE"})

			slaves := state.Slaves
			for _, s := range slaves {
				s.Type = "agent"
				s.Region = s.Domain.FaultDomain.Region.Name
				s.Zone = s.Domain.FaultDomain.Zone.Name
				s.PublicIPs = ips[s.IP()]
				nodes = append(nodes, s)
				table.Append([]string{s.Hostname, s.IP(), strings.Join(s.PublicIPs, ", "), s.ID, s.Type, s.Region, s.Zone})
			}

			for _, m := range masters {
				m.Type = "master"
				m.PublicIPs = ips[m.IP]
				if m.IP == state.Hostname {
					m.Type = "master (leader)"
					m.Region = state.Domain.FaultDomain.Region.Name
					m.Zone = state.Domain.FaultDomain.Zone.Name
					m.ID, m.PID, m.Version = state.ID, state.PID, state.Version
				}
				nodes = append(nodes, m)
				table.Append([]string{
					m.Host,
					m.IP,
					strings.Join(m.PublicIPs, ", "),
					tablewriter.ConditionString(m.ID != "", m.ID, "N/A"),
					m.Type,
					tablewriter.ConditionString(m.Region != "", m.Region, "N/A"),
					tablewriter.ConditionString(m.Zone != "", m.Zone, "N/A"),
				})
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(nodes)
			}

			table.Render()

			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	return cmd
}
