package debug

import (
	"encoding/json"
	"fmt"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/marathon"
	"github.com/spf13/cobra"
)

func newCmdMarathonDebugDetails(ctx api.Context) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "details",
		Short: "Display detailed information for a queued instance launch for debugging purpose.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := marathon.NewClient(ctx)
			if err != nil {
				return err
			}

			return marathonDebugDetails(ctx, client, args[0], jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print JSON-formatted data.")

	return cmd
}

func marathonDebugDetails(ctx api.Context, client *marathon.Client, id string, jsonOutput bool) error {
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
			table := cli.NewTable(ctx.Out(), []string{"HOSTNAME", "ROLE", "CONSTRAINTS", "CPUS", "MEM", "DISK", "PORTS", "SCARCE", "RECEIVED"})
			for _, unusedOffer := range item.LastUnusedOffers {
				reasons := unusedOffer.Reason
				table.Append([]string{unusedOffer.Offer.Hostname,
					isConstraintOK(marathon.UnfulfilledRole, reasons),
					isConstraintOK(marathon.UnfulfilledConstraint, reasons),
					isConstraintOK(marathon.InsufficientCpus, reasons),
					isConstraintOK(marathon.InsufficientMemory, reasons),
					isConstraintOK(marathon.InsufficientDisk, reasons),
					isConstraintOK(marathon.InsufficientPorts, reasons),
					isConstraintOK(marathon.DeclinedScarceResources, reasons),
					unusedOffer.Timestamp})
			}
			table.Render()
			return nil
		}
	}

	return fmt.Errorf("couldn't find app %s in Marathon queue", id)
}

func isConstraintOK(constraint string, reasons []string) string {
	for _, reason := range reasons {
		if reason == constraint {
			return "-"
		}
	}
	return "ok"
}
