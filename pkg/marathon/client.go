package marathon

import (
	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	marathon "github.com/gambol99/go-marathon"
)

// Client to interact with the Marathon API.
type Client struct {
	API marathon.Marathon
}

// NewClient creates a new HTTP wrapper client to talk to the Marathon service.
func NewClient(ctx api.Context) (*Client, error) {
	cluster, err := ctx.Cluster()
	if err != nil {
		return nil, err
	}
	baseURL, _ := cluster.Config().Get("marathon.url").(string)
	if baseURL == "" {
		baseURL = cluster.URL() + "/service/marathon"
	}

	dcosClient := pluginutil.NewHTTPClient(baseURL)

	config := marathon.NewDefaultConfig()
	config.URL = baseURL
	config.HTTPClient = dcosClient

	client, err := marathon.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &Client{API: client}, nil
}
