package diagnostics

import (
	"errors"
	"fmt"
	"time"

	"github.com/dcos/dcos-cli/api"
	diagnostics "github.com/dcos/dcos-core-cli/pkg/diagnostics/v2"
	"github.com/spf13/cobra"
)

// NewCommand creates and returns a diagnostics command with its subcommands
// already added.
func NewCommand(ctx api.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diagnostics",
		Short: "Create and manage DC/OS diagnostics bundles",
	}
	cmd.AddCommand(
		newDiagnosticsListCommand(ctx),
		newDiagnosticsDownloadCommand(ctx),
		newDiagnosticsCreateCommand(ctx),
		newDiagnosticsDeleteCommand(ctx),
		newDiagnosticsWaitCommand(ctx),
	)
	return cmd
}

func latestBundle(client *diagnostics.Client) (*diagnostics.Bundle, error) {
	bundles, err := client.List()
	if err != nil {
		return nil, err
	}
	if len(bundles) == 0 {
		return nil, errors.New("no bundles found")
	}

	// default time.Time is as far back we'll need to worry about anyway so serves as a good starting min
	var max time.Time
	var bundle diagnostics.Bundle

	for _, b := range bundles {
		if b.Type == diagnostics.Cluster && b.Started.After(max) {
			max = b.Started
			bundle = b
		}
	}

	if max.Equal(time.Time{}) {
		return nil, errors.New("no cluster bundles found")
	}

	return &bundle, nil
}

func client(ctx api.Context) (*diagnostics.Client, error) {
	cluster, err := ctx.Cluster()
	if err != nil {
		return nil, fmt.Errorf("could not get cluster: %s", err)
	}
	c, err := ctx.HTTPClient(cluster)
	if err != nil {
		return nil, fmt.Errorf("could not create HTTP client: %s", err)
	}
	return diagnostics.NewClient(c), nil
}
