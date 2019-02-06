package node

import (
	"encoding/json"
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/diagnostics"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

func newCmdNodeListComponents(ctx api.Context) *cobra.Command {
	var jsonOutput bool
	var leader bool
	var mesosID string
	cmd := &cobra.Command{
		Use:   "list-components",
		Short: "Print a list of available DC/OS components on specified node",
		Args:  cobra.NoArgs,
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
				leader, err := mesosClient().Leader()
				if err != nil {
					return err
				}
				if leader.IP == "" {
					return fmt.Errorf("invalid leader response, missing field 'ip'")
				}
				ip = leader.IP
			} else {
				agents, err := mesosClient().Agents()
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

			units, err := diagnosticsClient().Units(ip)
			if err != nil {
				return err
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(units.Array)
			}
			for _, component := range units.Array {
				fmt.Println(component.UnitID)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	cmd.Flags().BoolVar(&leader, "leader", false, "The leading master")
	cmd.Flags().StringVar(&mesosID, "mesos-id", "", "The agent ID of a node")
	return cmd
}
