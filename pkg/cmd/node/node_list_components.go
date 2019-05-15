package node

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/diagnostics"
	"github.com/spf13/cobra"
)

func newCmdNodeListComponents(ctx api.Context) *cobra.Command {
	var jsonOutput bool
	var leader bool
	var mesosID, unitType string
	cmd := &cobra.Command{
		Use:     "list-units",
		Aliases: []string{"list-components"},
		Short:   "Print a list of available DC/OS systemd units on specified node",
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !leader && mesosID == "" {
				return fmt.Errorf("'--leader' or '--mesos-id' must be provided")
			} else if leader && mesosID != "" {
				return fmt.Errorf("unable to use leader and mesos id at the same time")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ip := ""
			if leader {
				leader, err := mesosDNSClient().Leader()
				if err != nil {
					return err
				}
				if leader.IP == "" {
					return fmt.Errorf("invalid leader response, missing field 'ip'")
				}
				ip = leader.IP
			} else {
				c, err := mesosClient(ctx)
				if err != nil {
					return err
				}
				agents, err := c.Agents()
				if err != nil {
					return err
				}
				for _, agent := range agents {
					if mesosID == agent.AgentInfo.GetID().Value {
						ip = agent.AgentInfo.GetHostname()
					}
				}
				if ip == "" {
					return fmt.Errorf("agent '%s' not found", mesosID)
				}
			}

			unitsHealth, err := diagnosticsClient().Units(ip)
			if err != nil {
				return err
			}

			var units []diagnostics.HealthResponseValues

			for _, unit := range unitsHealth.Array {
				if unitType != "" && !strings.HasSuffix(unit.UnitID, "."+unitType) {
					continue
				}
				units = append(units, unit)
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(units)
			}
			for _, unit := range units {
				fmt.Println(strings.TrimSuffix(unit.UnitID, "."+unitType))
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	cmd.Flags().BoolVar(&leader, "leader", false, "The leading master")
	cmd.Flags().StringVar(&mesosID, "mesos-id", "", "The agent ID of a node")
	cmd.Flags().StringVar(&unitType, "type", "", "Only list a given type of unit (eg. service, socket, etc.)")
	return cmd
}
