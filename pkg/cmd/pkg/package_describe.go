package pkg

import (
	"encoding/base64"
	"encoding/json"

	"github.com/dcos/client-go/dcos"
	"github.com/dcos/dcos-cli/api"
	"github.com/spf13/cobra"

	"github.com/dcos/dcos-core-cli/pkg/cosmos"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
)

type describeOptions struct {
	appID       string
	optionsPath string
	version     string

	allVersions bool
	appOnly     bool
	cliOnly     bool
	config      bool
	render      bool
}

func newCmdPackageDescribe(ctx api.Context) *cobra.Command {
	var opts describeOptions
	cmd := &cobra.Command{
		Use:   "describe <package-name>",
		Short: "Get specific details for packages",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pkgDescribe(ctx, args[0], opts)
		},
	}
	cmd.Flags().BoolVar(&opts.appOnly, "app", false, "Application only")
	cmd.Flags().StringVar(&opts.appID, "app-id", "", "The application ID")
	cmd.Flags().BoolVar(&opts.cliOnly, "cli", false, "Command line only")
	cmd.Flags().BoolVar(&opts.config, "config", false, "Print the opts.configurable properties of the `marathon.json` file")
	cmd.Flags().StringVar(&opts.optionsPath, "options", "", "Path to a JSON file that contains customized package installation describeOptions")
	cmd.Flags().StringVar(&opts.version, "package-version", "", "The package opts.version")
	cmd.Flags().BoolVar(&opts.allVersions, "package-versions", false, " Displays all versions for this package")
	cmd.Flags().BoolVar(&opts.render, "render", false,
		"Collate the `marathon.json` package template with values from the `opts.config.json` and `--describeOptions`")
	return cmd
}

func pkgDescribe(ctx api.Context, packageName string, opts describeOptions) error {
	c, err := cosmos.NewClient(ctx, pluginutil.HTTPClient(""))
	if err != nil {
		return err
	}
	enc := json.NewEncoder(ctx.Out())
	enc.SetIndent("", "  ")
	if opts.allVersions {
		versions, err := c.PackageListVersions(packageName)
		if err != nil {
			return err
		}
		return enc.Encode(versions)

	}
	description, err := c.PackageDescribe(packageName, opts.version)
	if err != nil {
		return err
	}
	if opts.appOnly {
		// If the user supplied template describeOptions, they definitely want to render the template
		if opts.render || opts.optionsPath != "" {
			template, err := c.PackageRender(opts.appID, packageName, opts.version, opts.optionsPath)
			if err != nil {
				return err
			}
			return enc.Encode(template)
		}

		template, err := base64.StdEncoding.DecodeString(description.Package.Marathon.V2AppMustacheTemplate)
		if err != nil {
			return err
		}
		_, err = ctx.Out().Write(template)
		return err
	}
	if opts.cliOnly {
		if !isEmptyCli(description.Package.Resource.Cli) {
			return enc.Encode(description.Package.Resource.Cli)
		}
		if !isEmptyCommand(description.Package.Command) {
			return enc.Encode(description.Package.Command)
		}
		return nil
	}
	if opts.config {
		return enc.Encode(description.Package.Config)
	}
	return enc.Encode(description)
}

func isEmptyCommand(cmd dcos.CosmosPackageCommand) bool {
	return cmd.Name == "" && len(cmd.Pip) == 0
}

func isEmptyCli(cli dcos.CosmosPackageResourceCli) bool {
	return isEmptyBinary(cli.Binaries.Darwin) &&
		isEmptyBinary(cli.Binaries.Linux) &&
		isEmptyBinary(cli.Binaries.Windows)
}

func isEmptyBinary(binary dcos.CosmosPackageResourceCliOsBinaries) bool {
	x := binary.X8664
	return len(x.ContentHash) == 0 && x.Kind == "" && x.Url == ""
}
