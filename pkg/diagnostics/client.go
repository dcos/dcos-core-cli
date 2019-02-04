package diagnostics

import (
	"encoding/json"
	"fmt"

	"github.com/dcos/dcos-cli/pkg/httpclient"
)

// Client is a diagnostics client for DC/OS.
type Client struct {
	http *httpclient.Client
}

// NewClient creates a new diagnostics client.
func NewClient(baseClient *httpclient.Client) *Client {
	return &Client{
		http: baseClient,
	}
}

// Units returns the units of a certain node.
func (c *Client) Units(node string) (*UnitsHealthResponseJSONStruct, error) {
	resp, err := c.http.Get(fmt.Sprintf("/system/health/v1/nodes/%s/units", node))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		var units UnitsHealthResponseJSONStruct
		err = json.NewDecoder(resp.Body).Decode(&units)
		if err != nil {
			return nil, err
		}
		return &units, nil
	default:
		return nil, fmt.Errorf("HTTP %d error", resp.StatusCode)
	}
}
