package node

import (
	"encoding/json"
	"fmt"
	"reflect"
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

type stateSummaryResult struct {
	state *mesos.StateSummary
	err   error
}

type ipsResult map[string][]string

func newCmdNodeList(ctx api.Context) *cobra.Command {
	var jsonOutput bool
	var fields []string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Show all nodes in the cluster",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mesosClient(ctx)
			if err != nil {
				return err
			}

			// The following code allows us to do the API calls for the
			// state, the public IPs, and the masters concurrently which
			// results in a significant speedup of the command.
			stateRes := make(chan stateResult)
			go func() {
				state, err := client.State()
				stateRes <- stateResult{state, err}
			}()

			stateSummaryRes := make(chan stateSummaryResult)
			go func() {
				stateSummary, err := client.StateSummary()
				stateSummaryRes <- stateSummaryResult{stateSummary, err}
			}()

			ipsRes := make(chan ipsResult)
			go func() {
				c := networking.NewClient(pluginutil.HTTPClient(""))
				nodes, err := c.Nodes()

				ips := make(map[string][]string)
				if err != nil {
					ctx.Logger().Debug(err)
				} else {
					for _, node := range nodes {
						ips[node.PrivateIP] = node.PublicIPs
					}
				}
				ipsRes <- ips
			}()

			masters, err := mesosDNSClient().Masters()
			if err != nil {
				return err
			}

			ips := <-ipsRes

			stateSummaryResult := <-stateSummaryRes
			if stateSummaryResult.err != nil {
				return stateSummaryResult.err
			}
			stateSummary := stateSummaryResult.state

			stateResult := <-stateRes
			if stateResult.err != nil {
				return stateResult.err
			}
			state := stateResult.state

			// In order to create a nodes json object that contains masters and agents
			// we need a slice of interface{} that is able to contain both node types.
			nodes := make([]interface{}, 0)
			tableHeader := []string{"HOSTNAME", "IP", "PUBLIC IP(S)", "ID", "TYPE", "REGION", "ZONE"}
			for _, field := range fields {
				tableHeader = append(tableHeader, field)
			}
			table := cli.NewTable(ctx.Out(), tableHeader)

			slaves := state.Slaves
			for _, s := range slaves {
				s.Type = "agent"
				s.Region = s.Domain.FaultDomain.Region.Name
				s.Zone = s.Domain.FaultDomain.Zone.Name
				s.PublicIPs = ips[s.IP()]

				for _, sSummary := range stateSummary.Slaves {
					if sSummary.ID == s.ID {
						s.TaskError = sSummary.TaskError
						s.TaskFailed = sSummary.TaskFailed
						s.TaskFinished = sSummary.TaskFinished
						s.TaskKilled = sSummary.TaskKilled
						s.TaskKilling = sSummary.TaskKilling
						s.TaskLost = sSummary.TaskLost
						s.TaskRunning = sSummary.TaskRunning
						s.TaskStaging = sSummary.TaskStaging
						s.TaskStarting = sSummary.TaskStarting
					}
				}

				nodes = append(nodes, s)

				// Additional information, only for non JSON output.
				if val, ok := s.Attributes["public_ip"].(string); ok && val == "true" {
					s.Type = "agent (public)"
				}

				tableItem := []string{s.Hostname, s.IP(), strings.Join(s.PublicIPs, ", "), s.ID, s.Type, s.Region, s.Zone}
				for _, field := range fields {
					tableItem = append(tableItem, nodeExtraField(s, strings.Split(field, ".")))
				}
				table.Append(tableItem)
			}

			for _, m := range masters {
				m.Type = "master (standby)"
				// All masters must be in the same region:
				// https://github.com/apache/mesos/blob/3944124da5338791ce28c4a9285c98ee80c99b16/src/master/master.cpp#L2127
				m.Region = state.Domain.FaultDomain.Region.Name
				m.PublicIPs = ips[m.IP]
				if m.IP == state.Hostname {
					m.Type = "master (leader)"
					// The only way to get the zone for non-leader masters would be
					// to SSH on each of them and get the content of `GET_FLAGS`.
					m.Zone = state.Domain.FaultDomain.Zone.Name
					// Only the leader master has an ID:
					// https://github.com/apache/mesos/blob/3944124da5338791ce28c4a9285c98ee80c99b16/src/common/protobuf_utils.cpp#L565
					m.ID, m.PID, m.Version = state.ID, state.PID, state.Version
				}

				nodes = append(nodes, m)

				tableItem := []string{
					m.Host,
					m.IP,
					strings.Join(m.PublicIPs, ", "),
					tablewriter.ConditionString(m.ID != "", m.ID, "N/A"),
					m.Type,
					tablewriter.ConditionString(m.Region != "", m.Region, "N/A"),
					tablewriter.ConditionString(m.Zone != "", m.Zone, "N/A"),
				}
				for _, field := range fields {
					tableItem = append(tableItem, nodeExtraField(m, strings.Split(field, ".")))
				}
				table.Append(tableItem)
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
	cmd.Flags().StringArrayVar(&fields, "field", nil, "Name of extra field to include in the output of `dcos node`. Can be repeated multiple times to add several fields.")
	return cmd
}

func nodeExtraField(data interface{}, field []string) string {
	val := reflect.ValueOf(data)
	switch val.Kind() {
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			tag, ok := val.Type().Field(i).Tag.Lookup("json")
			tag = strings.Split(tag, ",")[0]
			if !ok || len(field) == 0 || tag != field[0] {
				continue
			}
			return nodeExtraField(val.Field(i).Interface(), field[1:])
		}
	default:
		return fmt.Sprintf("%v", val)
	}
	return ""
}
