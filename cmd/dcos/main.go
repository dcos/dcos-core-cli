package main

import (
	"fmt"
	"os"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-cli/pkg/httpclient"
	"github.com/dcos/dcos-core-cli/pkg/cmd"
)

func main() {
	ctx := cli.NewContext(cli.NewOsEnvironment())
	if err := run(ctx, os.Args); err != nil {
		fmt.Fprintf(ctx.ErrOut(), "Error: %s\n", errorMessage(err))
		os.Exit(1)
	}
}

func run(ctx api.Context, args []string) error {
	return cmd.NewDCOSCommand(ctx).Execute()
}

func errorMessage(err error) string {
	if httpErr, ok := err.(*httpclient.HTTPError); ok {
		switch httpErr.Response.StatusCode {
		case 401:
			return "authentication failed, please run `dcos auth login`"
		case 403:
			return "you are not authorized to perform this operation"
		}
	}
	return err.Error()
}
