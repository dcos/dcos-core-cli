package pkg

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dcos/client-go/dcos"
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/plugin"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/dcos/dcos-core-cli/pkg/cosmos"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
)

type pkgInstallOptions struct {
	appID       string
	optionsPath string
	version     string
	appOnly     bool
	cliOnly     bool
	yes         bool
}

func newCmdPackageInstall(ctx api.Context) *cobra.Command {
	var options pkgInstallOptions

	cmd := &cobra.Command{
		Use:   "install <package-name>",
		Short: "Install a package",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pkgInstall(ctx, args[0], options)
		},
	}
	cmd.Flags().BoolVar(&options.appOnly, "app", false, "Application only")
	cmd.Flags().StringVar(&options.appID, "app-id", "", "The application ID")
	cmd.Flags().BoolVar(&options.cliOnly, "cli", false, "Command line only")
	cmd.Flags().StringVar(&options.optionsPath, "options", "", "Path to a JSON file that contains customized package installation options")
	cmd.Flags().StringVar(&options.version, "package-version", "", "The package version")
	cmd.Flags().BoolVar(&options.yes, "yes", false, "Disable interactive mode and assume “yes” is the answer to all prompts")
	return cmd
}

func pkgInstall(ctx api.Context, packageName string, opts pkgInstallOptions) error {
	if packageName == "" {
		return fmt.Errorf("package name must not be empty")
	}

	if opts.cliOnly && opts.appOnly {
		return fmt.Errorf("--app and --cli are mutually exclusive")
	}
	// Install both if neither flag is specified
	if opts.cliOnly == opts.appOnly {
		opts.cliOnly = true
		opts.appOnly = true
	}
	if opts.optionsPath != "" {
		_, err := os.Stat(opts.optionsPath)
		if err != nil {
			return err
		}
	}
	c, err := cosmos.NewClient(ctx, pluginutil.HTTPClient(""))
	if err != nil {
		return err
	}
	description, err := c.PackageDescribe(packageName, opts.version)
	if err != nil {
		return err
	}
	pkg := description.Package
	link := "https://mesosphere.com/catalog-terms-conditions/#certified-services"
	if !pkg.Selected {
		fmt.Fprint(ctx.ErrOut(), "This is a Community service. "+
			"Community services are not tested for production environments. "+
			"There may be bugs, incomplete features, incorrect documentation, or other discrepancies.\n")
		link = "https://mesosphere.com/catalog-terms-conditions/#community-services"
	}
	fmt.Fprintf(ctx.ErrOut(), "By Deploying, you agree to the Terms and Conditions %s\n", link)
	if opts.appOnly && pkg.PreInstallNotes != "" {
		fmt.Fprintf(ctx.ErrOut(), "%s\n", pkg.PreInstallNotes)
	}
	if !opts.yes {
		err = ctx.Prompt().Confirm("Continue installing? [yes/no] ", "Yes")
		if err != nil {
			return err
		}
	}
	if opts.appOnly && description.Package.Marathon.V2AppMustacheTemplate != "" {
		fmt.Fprintf(ctx.ErrOut(), "Installing Marathon app for package [%s] version [%s]\n", packageName, pkg.Version)
		err := c.PackageInstall(opts.appID, packageName, pkg.Version, opts.optionsPath)
		if err != nil {
			return err
		}
	}
	if opts.cliOnly && isEmptyCli(pkg.Resource.Cli) && !isEmptyCommand(pkg.Command) {
		return fmt.Errorf("unable to install CLI subcommand. PIP subcommands are no longer supported" +
			" see: https://godoc.org/github.com/dcos/client-go/dcos#CosmosPackageCommand")
	}
	if opts.cliOnly && !isEmptyCli(pkg.Resource.Cli) {
		err := installCliPlugin(ctx, pkg)
		if err != nil {
			return err
		}
	}
	if opts.appOnly && pkg.PostInstallNotes != "" {
		fmt.Fprintf(ctx.ErrOut(), "%s\n", pkg.PostInstallNotes)
	}
	return nil
}

func installCliPlugin(ctx api.Context, pkg dcos.CosmosPackage) error {
	fmt.Fprintf(ctx.ErrOut(), "Installing CLI subcommand for package [%s] version [%s]\n", pkg.Name, pkg.Version)
	pluginInfo, err := cosmos.CLIPluginInfo(pkg, pluginutil.HTTPClient("").BaseURL())
	if err != nil {
		return fmt.Errorf("cannot get plugin info: %s", err)
	}
	var checksum plugin.Checksum
	for _, contentHash := range pluginInfo.ContentHash {
		switch contentHash.Algo {
		case dcos.SHA256:
			checksum.Hasher = sha256.New()
			checksum.Value = contentHash.Value
		default:
			fmt.Fprintf(ctx.ErrOut(), "unknown algorithm: %s\n", contentHash.Algo)
		}
	}
	cluster, err := ctx.Cluster()
	if err != nil {
		return fmt.Errorf("cannot get cluster: %s", err)
	}
	p, err := ctx.PluginManager(cluster).Install(pluginInfo.Url, &plugin.InstallOpts{
		Name:     pkg.Name,
		Update:   true,
		Checksum: checksum,
		PostInstall: func(fs afero.Fs, pluginDir string) error {
			pkgInfoFilepath := filepath.Join(pluginDir, "package.json")
			pkgInfoFile, err := fs.OpenFile(pkgInfoFilepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			defer pkgInfoFile.Close()
			return json.NewEncoder(pkgInfoFile).Encode(pkg)
		},
	})
	if err != nil {
		return fmt.Errorf("cannot install plugin: %s", err)
	}

	fmt.Fprintf(ctx.ErrOut(), "New commands available: %s\n", strings.Join(p.CommandNames(), ", "))
	return nil
}
