package debug

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/marathon"
	goMarathon "github.com/gambol99/go-marathon"
	"github.com/spf13/cobra"
)

func newCmdMarathonDebugSummary(ctx api.Context) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Display summarized information for a queued instance launch for debugging purpose.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := marathon.NewClient(ctx)
			if err != nil {
				return err
			}

			return marathonDebugSummary(ctx, client, args[0], jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print JSON-formatted data.")

	return cmd
}

const nilString = "---"

func marathonDebugSummary(ctx api.Context, client *marathon.Client, id string, jsonOutput bool) error {
	if jsonOutput {
		rawQueue, err := client.RawQueue()
		if err != nil {
			return err
		}

		for _, el := range rawQueue.Queue {
			app, okApp := el["app"].(map[string]interface{})
			pod, okPod := el["pod"].(map[string]interface{})
			if (okApp && app["id"] == id) || (okPod && pod["id"] == id) {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(el)
			}
		}
		return nil
	}

	queue, err := client.QueueWithLastUnusedOffers()
	if err != nil {
		return err
	}

	for _, item := range queue.Items {
		if (item.Application != nil && item.Application.ID == id) || (item.Pod != nil && item.Pod.ID == id) {
			agents := len(item.LastUnusedOffers)
			r := make(map[string]int)
			r[marathon.UnfulfilledRole] = agents
			r[marathon.UnfulfilledConstraint] = agents
			r[marathon.InsufficientCpus] = agents
			r[marathon.InsufficientMemory] = agents
			r[marathon.InsufficientDisk] = agents
			r[marathon.InsufficientPorts] = agents

			for _, unusedOffer := range item.LastUnusedOffers {
				for _, reason := range unusedOffer.Reason {
					if _, ok := r[reason]; ok {
						r[reason]--
					}
				}
			}

			table := cli.NewTable(ctx.Out(), []string{"RESOURCE", "REQUESTED", "MATCHED", "PERCENTAGE"})

			if item.Application != nil {
				table.Append([]string{
					"ROLE",
					nilPrintString(item.Application.Role),
					fmt.Sprintf("%d/%d", r[marathon.UnfulfilledRole], agents),
					fmt.Sprintf("%.2f%%", float64(r[marathon.UnfulfilledRole])/float64(agents)*100)})
				table.Append([]string{
					"CONSTRAINTS",
					nilPrintSliceOfSliceOfString(item.Application.Constraints),
					fmt.Sprintf("%d/%d", r[marathon.UnfulfilledConstraint], agents),
					fmt.Sprintf("%.2f%%", float64(r[marathon.UnfulfilledConstraint])/float64(agents)*100)})
				table.Append([]string{
					"CPUS",
					fmt.Sprintf("%v", item.Application.CPUs),
					fmt.Sprintf("%d/%d", r[marathon.InsufficientCpus], agents),
					fmt.Sprintf("%.2f%%", float64(r[marathon.InsufficientCpus])/float64(agents)*100)})
				table.Append([]string{
					"MEM",
					nilPrintFloat64(item.Application.Mem),
					fmt.Sprintf("%d/%d", r[marathon.InsufficientMemory], agents),
					fmt.Sprintf("%.2f%%", float64(r[marathon.InsufficientMemory])/float64(agents)*100)})
				table.Append([]string{
					"DISK",
					nilPrintFloat64(item.Application.Disk),
					fmt.Sprintf("%d/%d", r[marathon.InsufficientDisk], agents),
					fmt.Sprintf("%.2f%%", float64(r[marathon.InsufficientDisk])/float64(agents)*100)})
				table.Append([]string{
					"PORTS",
					fmt.Sprintf("%v", item.Application.Ports),
					fmt.Sprintf("%d/%d", r[marathon.InsufficientPorts], agents),
					fmt.Sprintf("%.2f%%", float64(r[marathon.InsufficientPorts])/float64(agents)*100)})
			} else if item.Pod != nil {
				table.Append([]string{
					"ROLE",
					nilPrintString(item.Pod.Role),
					fmt.Sprintf("%d/%d", r[marathon.UnfulfilledRole], agents),
					fmt.Sprintf("%.2f%%", float64(r[marathon.UnfulfilledRole])/float64(agents)*100)})
				table.Append([]string{
					"CONSTRAINTS",
					nilPrintSliceOfConstraints(item.Pod.Scheduling.Placement.Constraints),
					fmt.Sprintf("%d/%d", r[marathon.UnfulfilledConstraint], agents),
					fmt.Sprintf("%.2f%%", float64(r[marathon.UnfulfilledConstraint])/float64(agents)*100)})
				table.Append([]string{
					"CPUS",
					fmt.Sprintf("%v", item.Pod.ExecutorResources.Cpus),
					fmt.Sprintf("%d/%d", r[marathon.InsufficientCpus], agents),
					fmt.Sprintf("%.2f%%", float64(r[marathon.InsufficientCpus])/float64(agents)*100)})
				table.Append([]string{
					"MEM",
					fmt.Sprintf("%v", item.Pod.ExecutorResources.Mem),
					fmt.Sprintf("%d/%d", r[marathon.InsufficientMemory], agents),
					fmt.Sprintf("%.2f%%", float64(r[marathon.InsufficientMemory])/float64(agents)*100)})
				table.Append([]string{
					"DISK",
					fmt.Sprintf("%v", item.Pod.ExecutorResources.Disk),
					fmt.Sprintf("%d/%d", r[marathon.InsufficientDisk], agents),
					fmt.Sprintf("%.2f%%", float64(r[marathon.InsufficientDisk])/float64(agents)*100)})
				table.Append([]string{
					"PORTS",
					nilString,
					fmt.Sprintf("%d/%d", r[marathon.InsufficientPorts], agents),
					fmt.Sprintf("%.2f%%", float64(r[marathon.InsufficientPorts])/float64(agents)*100)})
			}
			table.Render()
			return nil
		}
	}

	return errors.New("no apps found in Marathon queue")
}

func nilPrintString(val *string) string {
	if val != nil {
		return *val
	}
	return nilString
}

func nilPrintSliceOfConstraints(val *[]goMarathon.Constraint) string {
	if val != nil {
		return fmt.Sprintf("%v", *val)
	}
	return nilString
}

func nilPrintSliceOfSliceOfString(val *[][]string) string {
	if val != nil {
		return fmt.Sprintf("%v", *val)
	}
	return nilString
}

func nilPrintFloat64(val *float64) string {
	if val != nil {
		return fmt.Sprintf("%v", *val)
	}
	return nilString
}
