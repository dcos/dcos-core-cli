package metrics

import (
	"encoding/json"
	"fmt"

	"github.com/dcos/dcos-cli/pkg/httpclient"
)

// Client is a metrics client for DC/OS.
type Client struct {
	http *httpclient.Client
}

// NewClient creates a new metrics client.
func NewClient(baseClient *httpclient.Client) *Client {
	return &Client{
		http: baseClient,
	}
}

// Node returns the units of a certain node.
func (c *Client) Node(mesosID string) (*Node, error) {
	resp, err := c.http.Get(fmt.Sprintf("/system/v1/agent/%s/metrics/v0/node", mesosID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		var node Node
		err = json.NewDecoder(resp.Body).Decode(&node)
		if err != nil {
			return nil, err
		}
		return &node, nil
	default:
		return nil, fmt.Errorf("HTTP %d error", resp.StatusCode)
	}
}
