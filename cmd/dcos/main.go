package main

import (
	"crypto/x509"
	"fmt"
	"net/url"
	"os"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/cli"
	"github.com/dcos/dcos-cli/pkg/httpclient"
	"github.com/dcos/dcos-core-cli/pkg/cmd"
)

const invalidCertError = "An SSL error occurred. To configure your SSL settings, " +
	"please run: 'dcos config set core.ssl_verify <value>'\n" +
	"<value>: Whether to verify SSL certs for HTTPS or path to certs. " +
	"Valid values are a path to a CA_BUNDLE, True (will then use system CA " +
	"certificates), or False (will then send insecure requests)."

func main() {
	ctx := cli.NewContext(cli.NewOsEnvironment())
	if err := run(ctx); err != nil {
		fmt.Fprintf(ctx.ErrOut(), "Error: %s\n", errorMessage(err))
		os.Exit(1)
	}
}

func run(ctx api.Context) error {
	return cmd.NewDCOSCommand(ctx).Execute()
}

func errorMessage(err error) string {
	switch e := err.(type) {
	case *httpclient.HTTPError:
		switch e.Response.StatusCode {
		case 401:
			return "authentication failed, please run `dcos auth login`"
		case 403:
			return "you are not authorized to perform this operation"
		}
	case *url.Error:
		switch e.Err.(type) {
		case x509.CertificateInvalidError, x509.UnknownAuthorityError:
			return invalidCertError
		}
	}
	return err.Error()
}
