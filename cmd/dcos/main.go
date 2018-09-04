package main

import (
	"os"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-core-cli/pkg/cmd"
)

func main() {
	ctx := cli.NewContext(cli.NewOsEnvironment())
	if err := run(ctx, os.Args); err != nil {
		os.Exit(1)
	}
}

func run(ctx api.Context, args []string) error {
	return cmd.NewDCOSCommand(ctx).Execute()
}
