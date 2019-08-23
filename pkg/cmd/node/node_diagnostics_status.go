package node

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdNodeDiagnosticsStatus(ctx api.Context) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Print diagnostics job status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := ctx.Deprecated("This command is deprecated since DC/OS 1.14, please use 'dcos diagnostics list' instead.")
			if err != nil {
				return err
			}

			status, err := diagnosticsClient().Status()
			if err != nil {
				return err
			}

			if jsonOutput {
				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(status)
			}
			for k, v := range status {
				fmt.Println(k)
				val := reflect.ValueOf(v)
				t := val.Type()
				for i := 0; i < t.NumField(); i++ {
					fmt.Printf("  %s: ", t.Field(i).Tag.Get("json"))
					fmt.Println(val.Field(i).Interface())
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	return cmd
}
