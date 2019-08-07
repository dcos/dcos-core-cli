package cosmos

import (
	"context"

	"github.com/antihax/optional"
	"github.com/dcos/client-go/dcos"
	"github.com/dcos/dcos-cli/pkg/httpclient"
)

// Client is a diagnostics client for DC/OS.
type Client struct {
	cosmos *dcos.CosmosApiService
}

// NewClient creates a new diagnostics client.
func NewClient(baseClient *httpclient.Client) (*Client, error) {
	dcosClient, err := dcos.NewClient()
	if err != nil {
		return nil, err
	}
	return &Client{
		cosmos: dcosClient.Cosmos,
	}, nil
}

// PackageDescribe returns the content of '/package/describe'.
func (c *Client) PackageDescribe(name string, version string) (*Description, error) {
	desc, _, err := c.cosmos.PackageDescribe(context.TODO(), &dcos.PackageDescribeOpts{
		CosmosPackageDescribeV1Request: optional.NewInterface(dcos.CosmosPackageDescribeV1Request{
			PackageName:    name,
			PackageVersion: version,
		}),
	})
	if err != nil {
		return nil, err
	}

	backwardCompatibleDesc := Description{
		Package: desc.Package,
	}
	return &backwardCompatibleDesc, nil
}
