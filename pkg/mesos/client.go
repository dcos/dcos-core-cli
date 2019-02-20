package mesos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/dcos/dcos-cli/pkg/httpclient"
	"github.com/golang/protobuf/proto"
	mesos "github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/master"
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

// Hosts returns the IP address(es) of an host.
func (c *Client) Hosts(host string) ([]Host, error) {
	resp, err := c.http.Get("/mesos_dns/v1/hosts/" + host)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		var hosts []Host
		err = json.NewDecoder(resp.Body).Decode(&hosts)
		return hosts, err
	default:
		return nil, fmt.Errorf("HTTP %d error", resp.StatusCode)
	}
}

// Leader returns the Mesos leader of the connected cluster.
func (c *Client) Leader() (*Master, error) {
	resp, err := c.http.Get("/mesos_dns/v1/hosts/leader.mesos")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		var hosts []Master
		err = json.NewDecoder(resp.Body).Decode(&hosts)
		if len(hosts) > 1 {
			return nil, fmt.Errorf("expecting one leader. Got %d", len(hosts))
		}
		return &hosts[0], err
	default:
		return nil, fmt.Errorf("HTTP %d error", resp.StatusCode)
	}
}

// Masters returns the Mesos masters of the connected cluster.
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

// State returns the current State of the Mesos master.
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

// StateSummary returns a StateSummary of the Mesos master.
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

// Agents returns the agents of the mesos cluster.
func (c *Client) Agents() ([]master.Response_GetAgents_Agent, error) {
	body := master.Call{
		Type: master.Call_GET_AGENTS,
	}
	reqBody, err := proto.Marshal(&body)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Post("/mesos/api/v1", "application/x-protobuf", bytes.NewBuffer(reqBody),
		httpclient.Header("Accept", "application/x-protobuf"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var agents master.Response
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		err = proto.Unmarshal(bodyBytes, &agents)
		return agents.GetAgents.Agents, err
	case 503:
		return nil, fmt.Errorf("could not connect to the leading mesos master")
	default:
		return nil, fmt.Errorf("HTTP %d error", resp.StatusCode)
	}
}

// MarkAgentGone marks an agent as gone.
func (c *Client) MarkAgentGone(agentID string) error {
	body := master.Call{
		Type: master.Call_MARK_AGENT_GONE,
		MarkAgentGone: &master.Call_MarkAgentGone{
			AgentID: mesos.AgentID{Value: agentID},
		},
	}
	var reqBody bytes.Buffer
	if err := json.NewEncoder(&reqBody).Encode(body); err != nil {
		return err
	}
	resp, err := c.http.Post("/mesos/api/v1", "application/json", &reqBody, httpclient.FailOnErrStatus(false))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return fmt.Errorf("could not mark agent '%s' as gone", agentID)
	}
	return nil
}
