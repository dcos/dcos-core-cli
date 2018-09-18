package job

import (
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/metronome"
	"github.com/spf13/cobra"
)

// newCmdJobSchedule creates the `job schedule` command with all its subcommands.
func newCmdJobSchedule(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schedule",
		Short: "Managing schedules of jobs",
	}

	cmd.AddCommand(
		newCmdJobScheduleAdd(ctx),
		newCmdJobScheduleRemove(ctx),
		newCmdJobScheduleShow(ctx),
		newCmdJobScheduleUpdate(ctx),
	)

	return cmd
}

func parseJSONSchedule(r io.Reader) (*metronome.Schedule, error) {
	jsonBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var schedule metronome.Schedule
	if err := json.Unmarshal(jsonBytes, &schedule); err != nil {
		return nil, err
	}

	return &schedule, nil
}
