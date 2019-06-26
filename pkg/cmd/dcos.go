package cmd

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/cmd/job"
	"github.com/dcos/dcos-core-cli/pkg/cmd/node"
	"github.com/spf13/cobra"
)

const annotationUsageOptions string = "usage_options"

// NewDCOSCommand creates the `dcos` command with all the available subcommands.
func NewDCOSCommand(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use: "dcos",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmd.SilenceUsage = true
		},
	}

	cmd.AddCommand(
		job.NewCommand(ctx),
		node.NewCommand(ctx),
	)

	// This follows the CLI design guidelines for help formatting.
	cmd.SetUsageTemplate(`Usage:{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{else if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{.Name}}
      {{.Short}}{{end}}{{end}}{{end}}{{if or .HasAvailableLocalFlags (ne (index .Annotations "` + annotationUsageOptions + `") "")}}

Options:{{if ne (index .Annotations "` + annotationUsageOptions + `") ""}}{{index .Annotations "` + annotationUsageOptions + `"}}{{else}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)

	return cmd
}
