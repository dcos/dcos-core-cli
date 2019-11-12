package app

import (
	"encoding/json"
	"fmt"

	"github.com/dcos/dcos-core-cli/pkg/marathon"

	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"
)

func newCmdMarathonAppAdd(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add an application.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := marathon.NewClient(ctx)
			if err != nil {
				return err
			}
			if len(args) > 0 {
				return marathonAppAdd(ctx, *client, args[0])
			}
			return marathonAppAdd(ctx, *client, "")
		},
	}
	return cmd
}

func marathonAppAdd(ctx api.Context, client marathon.Client, appFile string) error {
	app, err := client.AddApp(ctx, appFile)
	if err != nil {
		if err == marathon.ErrCannotReadAppDefinition {
			return fmt.Errorf("can't read from resource: %s. Please check that it exists", appFile)
		} else if syntaxError, ok := err.(*json.SyntaxError); ok {
			return fmt.Errorf("error loading JSON: %s", syntaxError.Error())
		}
		return err
	}
	fmt.Fprintf(ctx.Out(), "Created deployment %s\n", app.Deployments[0]["id"])
	return nil
}
