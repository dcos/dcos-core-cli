package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/cosmos"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/spf13/cobra"
)

func newCmdPackageList(ctx api.Context) *cobra.Command {
	var cliOnly, jsonOutput bool
	var appID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Print a list of the installed DC/OS packages",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := cosmos.NewClient(ctx, pluginutil.HTTPClient(""))
			if err != nil {
				return err
			}

			listResult, err := c.PackageList()
			if err != nil {
				return err
			}
			list := *listResult

			if appID != "" {
				j := 0
				for _, p := range list {
					if p.Name == appID {
						list[j] = p
						j++
					}
				}
				list = list[:j]
			}

			if cliOnly {
				cluster, err := ctx.Cluster()
				if err != nil {
					return err
				}

				commands := make(map[string]bool)
				plugins := ctx.PluginManager(cluster).Plugins()
				for _, plugin := range plugins {
					for _, command := range plugin.Commands {
						commands[command.Name] = true
					}
				}

				k := 0
				for _, p := range list {
					if commands[p.Name] {
						list[k] = p
						k++
					}
				}
				list = list[:k]
			}

			if jsonOutput {
				if len(list) == 0 {
					fmt.Fprintln(ctx.Out(), "[]")
					return nil
				}

				enc := json.NewEncoder(ctx.Out())
				enc.SetIndent("", "    ")
				return enc.Encode(&list)
			}

			if len(list) == 0 {
				return errors.New("there are currently no installed packages")
			}

			table := cli.NewTable(ctx.Out(), []string{"NAME", "VERSION", "APP", "COMMAND", "DESCRIPTION"})
			for _, p := range list {
				command := "---"
				if p.Command != "" {
					command = p.Command
				}
				table.Append([]string{p.Name, p.Version, strings.Join(p.Apps, ", "), command, p.Description})
			}
			table.Render()

			return nil
		},
	}
	cmd.Flags().StringVar(&appID, "app-id", "", "The application ID")
	cmd.Flags().BoolVar(&cliOnly, "cli", false, "Command line only")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print in json format")
	return cmd
}
