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

// Node returns the metrics of a certain node.
func (c *Client) Node(agent string) (*Node, error) {
	resp, err := c.http.Get(fmt.Sprintf("/system/v1/agent/%s/metrics/v0/node", agent))
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

// Task returns the metrics of a certain task.
func (c *Client) Task(agent string, container string) (*Container, error) {
	resp, err := c.http.Get(fmt.Sprintf("/system/v1/agent/%s/metrics/v0/containers/%s", agent, container))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		var container Container
		err = json.NewDecoder(resp.Body).Decode(&container)
		if err != nil {
			return nil, err
		}
		return &container, nil
	case 204:
		return nil, nil
	default:
		return nil, fmt.Errorf("HTTP %d error", resp.StatusCode)
	}
}

// App returns the metrics of a certain app.
func (c *Client) App(agent string, container string) (*Container, error) {
	resp, err := c.http.Get(fmt.Sprintf("/system/v1/agent/%s/metrics/v0/containers/%s/app", agent, container))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		var container Container
		err = json.NewDecoder(resp.Body).Decode(&container)
		if err != nil {
			return nil, err
		}
		return &container, nil
	case 204:
		return nil, nil
	default:
		return nil, fmt.Errorf("HTTP %d error", resp.StatusCode)
	}
}
