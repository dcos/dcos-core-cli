package mesos

import (
	"encoding/json"
	"fmt"

	"github.com/dcos/dcos-cli/pkg/httpclient"
)

// Client is a Mesos client for DC/OS.
type Client struct {
	http *httpclient.Client
}

// NewClient creates a new Mesos client.
func NewClient(baseClient *httpclient.Client) *Client {
	return &Client{
		http: baseClient,
	}
}

// Masters is returning the mesos masters of the connected cluster.
func (c *Client) Masters() ([]Master, error) {
	resp, err := c.http.Get("/mesos_dns/v1/hosts/master.mesos")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		var hosts []Master
		err = json.NewDecoder(resp.Body).Decode(&hosts)
		return hosts, err
	default:
		return nil, fmt.Errorf("HTTP %d error", resp.StatusCode)
	}
}

// State returns the current State of the mesos master.
func (c *Client) State() (*State, error) {
	resp, err := c.http.Get("/mesos/master/state")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var state State
		err = json.NewDecoder(resp.Body).Decode(&state)
		return &state, err
	case 503:
		return nil, fmt.Errorf("could not connect to the leading mesos master")
	default:
		return nil, fmt.Errorf("HTTP %d error", resp.StatusCode)
	}
}

// StateSummary returns a StateSummary of the mesos master.
func (c *Client) StateSummary() (*StateSummary, error) {
	resp, err := c.http.Get("/mesos/master/state-summary")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var summary StateSummary
		err = json.NewDecoder(resp.Body).Decode(&summary)
		return &summary, err
	case 503:
		return nil, fmt.Errorf("could not connect to the leading mesos master")
	default:
		return nil, fmt.Errorf("HTTP %d error", resp.StatusCode)
	}
}

// Slaves returns the agents of the mesos cluster.
func (c *Client) Slaves() ([]Slave, error) {
	state, err := c.StateSummary()
	if err != nil {
		return nil, err
	}

	return state.Slaves, nil
}
