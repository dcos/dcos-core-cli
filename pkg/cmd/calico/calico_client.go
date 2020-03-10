package calico

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/httpclient"
	"github.com/dcos/dcos-core-cli/pkg/mesos"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
)

// Client is a calico client for DC/OS.
type Client struct {
	http *httpclient.Client
}

// NewClient creates a new calico client.
func NewClient(baseClient *httpclient.Client) *Client {
	return &Client{
		http: baseClient,
	}
}

// getCaCertificate gets the CA certificate from the server.
func (c *Client) getCaCertificate() (string, error) {
	resp, err := c.http.Get("/ca/dcos-ca.crt")
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return buf.String(), nil
	default:
		return "", httpResponseToError(resp)
	}
}

func httpResponseToError(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
	return &httpclient.HTTPError{
		Response: resp,
	}
}

func mesosClient(ctx api.Context) (*mesos.Client, error) {
	cluster, err := ctx.Cluster()
	if err != nil {
		return nil, err
	}
	baseURL, _ := cluster.Config().Get("core.mesos_master_url").(string)
	if baseURL == "" {
		baseURL = cluster.URL() + "/mesos"
	}
	return mesos.NewClient(pluginutil.HTTPClient(baseURL)), nil
}
