package debug

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/marathon"
	goMarathon "github.com/gambol99/go-marathon"
	"github.com/spf13/cobra"
)

const emptyEntry = "---"

func newCmdMarathonDebugList(ctx api.Context) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Print a list of currently queued instance launches for debugging purpose.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return debugList(ctx, jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print JSON-formatted data.")

	return cmd
}

func debugList(ctx api.Context, jsonOutput bool) error {
	client, err := marathon.NewClient(ctx)
	if err != nil {
		return err
	}

	queue, err := client.API.Queue()
	if err != nil {
		return err
	}

	if jsonOutput {
		enc := json.NewEncoder(ctx.Out())
		enc.SetIndent("", "    ")

		return enc.Encode(&queue.Items)

	}

	tableHeader := []string{"ID", "SINCE", "INSTANCES TO LAUNCH", "WAITING", "PROCESSED OFFERS",
		"UNUSED OFFERS", "LAST UNUSED OFFER", "LAST USED OFFER"}
	table := cli.NewTable(ctx.Out(), tableHeader)

	items := make([][]string, 0, len(queue.Items))

	for _, i := range queue.Items {
		items = append(items, []string{
			getAppOrPodID(i),
			i.Since,
			strconv.Itoa(i.Count),
			strconv.FormatBool(i.Delay.Overdue),
			strconv.Itoa(int(i.ProcessedOffersSummary.ProcessedOffersCount)),
			strconv.Itoa(int(i.ProcessedOffersSummary.UnusedOffersCount)),
			formatString(i.ProcessedOffersSummary.LastUnusedOfferAt),
			formatString(i.ProcessedOffersSummary.LastUsedOfferAt),
		})
	}

	sort.Sort(appRows(items))
	table.AppendBulk(items)
	table.Render()

	return nil
}

func getAppOrPodID(item goMarathon.Item) string {
	if item.Application != nil {
		return item.Application.ID
	}
	if item.Pod != nil {
		return item.Pod.ID
	}
	return ""
}

func formatString(s *string) string {
	if s != nil {
		return *s
	}
	return emptyEntry
}

// this enables sorting by ID
type appRows [][]string

func (a appRows) Swap(i int, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a appRows) Less(i int, j int) bool {
	return strings.Compare(a[i][0], a[j][0]) < 0
}

func (a appRows) Len() int {
	return len(a)
}
