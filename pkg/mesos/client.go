package mesos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/httpclient"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	google_protobuf "github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/proto"
	mesos "github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/master"
	"github.com/mesos/mesos-go/api/v1/lib/quota"
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

// NewClientWithContext returns a client with a `baseURL` to communicate with Mesos.
func NewClientWithContext(ctx api.Context) (*Client, error) {
	cluster, err := ctx.Cluster()
	if err != nil {
		return nil, err
	}
	baseURL, _ := cluster.Config().Get("core.mesos_master_url").(string)
	if baseURL == "" {
		baseURL = cluster.URL() + "/mesos"
	}
	return NewClient(pluginutil.HTTPClient(baseURL)), nil
}

// Debug returns the agent's internal virtual path mapping.
func (c *Client) Debug(agent string) (map[string]string, error) {
	resp, err := c.http.Get("/agent/" + agent + "/files/debug")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		debug := make(map[string]string)
		err = json.NewDecoder(resp.Body).Decode(&debug)
		if err != nil {
			return nil, err
		}

		return debug, nil
	default:
		return nil, httpResponseToError(resp)
	}
}

// Browse returns a file listing for an agent's directory
func (c *Client) Browse(agent string, path string) ([]File, error) {
	resp, err := c.http.Get("/agent/" + agent + "/files/browse?path=" + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		var browse []File
		err = json.NewDecoder(resp.Body).Decode(&browse)
		if err != nil {
			return nil, err
		}

		return browse, nil
	default:
		return nil, httpResponseToError(resp)
	}
}

// Download returns bytes read from a file in the sandbox of a task
// at the location of filePath.
func (c *Client) Download(agent string, filePath string) ([]byte, error) {
	resp, err := c.http.Get("/agent/" + agent + "/files/download?path=" + filePath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return b, nil
	default:
		return nil, httpResponseToError(resp)
	}
}

// Frameworks returns the frameworks of the connected cluster.
func (c *Client) Frameworks() ([]master.Response_GetFrameworks_Framework, error) {
	body := master.Call{
		Type: master.Call_GET_FRAMEWORKS,
	}
	reqBody, err := proto.Marshal(&body)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Post("/api/v1", "application/x-protobuf", bytes.NewBuffer(reqBody),
		httpclient.Header("Accept", "application/x-protobuf"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var frameworks master.Response
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		err = proto.Unmarshal(bodyBytes, &frameworks)
		return frameworks.GetFrameworks.Frameworks, err
	case 503:
		return nil, fmt.Errorf("could not connect to the leading mesos master")
	default:
		return nil, httpResponseToError(resp)
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
		return nil, httpResponseToError(resp)
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
		return nil, httpResponseToError(resp)
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
		return nil, httpResponseToError(resp)
	}
}

// Tasks returns all the tasks known in a Mesos cluster.
func (c *Client) Tasks() ([]mesos.Task, error) {
	body := master.Call{
		Type: master.Call_GET_TASKS,
	}
	reqBody, err := proto.Marshal(&body)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Post("/api/v1", "application/x-protobuf", bytes.NewBuffer(reqBody),
		httpclient.Header("Accept", "application/x-protobuf"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var tasks master.Response
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		err = proto.Unmarshal(bodyBytes, &tasks)
		allTasks := append(tasks.GetTasks.Tasks, tasks.GetTasks.CompletedTasks...)
		allTasks = append(allTasks, tasks.GetTasks.UnreachableTasks...)
		allTasks = append(allTasks, tasks.GetTasks.OrphanTasks...)
		return allTasks, err
	case 503:
		return nil, fmt.Errorf("could not connect to the leading mesos master")
	default:
		return nil, httpResponseToError(resp)
	}
}

// State returns the current State of the Mesos master.
func (c *Client) State() (*State, error) {
	resp, err := c.http.Get("/master/state")
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
		return nil, httpResponseToError(resp)
	}
}

// StateSummary returns a StateSummary of the Mesos master.
func (c *Client) StateSummary() (*StateSummary, error) {
	resp, err := c.http.Get("/master/state-summary")
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

	resp, err := c.http.Post("/api/v1", "application/x-protobuf", bytes.NewBuffer(reqBody),
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
		return nil, httpResponseToError(resp)
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
	resp, err := c.http.Post("/api/v1", "application/json", &reqBody, httpclient.FailOnErrStatus(false))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return nil
	case 404:
		return fmt.Errorf("could not mark agent '%s' as gone", agentID)
	default:
		return httpResponseToError(resp)
	}
}

// DeactivateAgent deactivates an agent.
func (c *Client) DeactivateAgent(agentID string) error {
	body := master.Call{
		Type: master.Call_DEACTIVATE_AGENT,
		DeactivateAgent: &master.Call_DeactivateAgent{
			AgentID: mesos.AgentID{Value: agentID},
		},
	}
	var reqBody bytes.Buffer
	if err := json.NewEncoder(&reqBody).Encode(body); err != nil {
		return err
	}
	resp, err := c.http.Post("/api/v1", "application/json", &reqBody, httpclient.FailOnErrStatus(false))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return nil
	case 404:
		return fmt.Errorf("could not deactivate agent '%s'", agentID)
	default:
		return httpResponseToError(resp)
	}
}

// ReactivateAgent reactivates an agent.
func (c *Client) ReactivateAgent(agentID string) error {
	body := master.Call{
		Type: master.Call_REACTIVATE_AGENT,
		ReactivateAgent: &master.Call_ReactivateAgent{
			AgentID: mesos.AgentID{Value: agentID},
		},
	}
	var reqBody bytes.Buffer
	if err := json.NewEncoder(&reqBody).Encode(body); err != nil {
		return err
	}
	resp, err := c.http.Post("/api/v1", "application/json", &reqBody, httpclient.FailOnErrStatus(false))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return nil
	case 404:
		return fmt.Errorf("could not reactivate agent '%s'", agentID)
	default:
		return httpResponseToError(resp)
	}
}

// DrainAgent drains an agent.
func (c *Client) DrainAgent(agentID string, gracePeriod time.Duration, markGone bool) error {
	body := master.Call{
		Type: master.Call_DRAIN_AGENT,
		DrainAgent: &master.Call_DrainAgent{
			AgentID:  mesos.AgentID{Value: agentID},
			MarkGone: &markGone,
		},
	}

	if gracePeriod != 0 {
		// As done in github.com/gogo/protobuf/types/duration.go.
		nanos := gracePeriod.Nanoseconds()
		secs := nanos / 1e9
		nanos -= secs * 1e9
		body.DrainAgent.MaxGracePeriod = &google_protobuf.Duration{
			Seconds: secs,
			Nanos:   int32(nanos),
		}
	}

	var reqBody bytes.Buffer
	if err := json.NewEncoder(&reqBody).Encode(body); err != nil {
		return err
	}
	resp, err := c.http.Post("/api/v1", "application/json", &reqBody, httpclient.FailOnErrStatus(false))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return nil
	case 404:
		return fmt.Errorf("could not drain agent '%s'", agentID)
	default:
		return httpResponseToError(resp)
	}
}

// Quota returns a quota.
func (c *Client) Quota() (*master.Response_GetQuota, error) {
	body := master.Call{
		Type: master.Call_GET_QUOTA,
	}
	reqBody, err := proto.Marshal(&body)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Post("/api/v1", "application/x-protobuf", bytes.NewBuffer(reqBody),
		httpclient.Header("Accept", "application/x-protobuf"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var response master.Response
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		err = proto.Unmarshal(bodyBytes, &response)
		return response.GetQuota, err
	case 503:
		return nil, fmt.Errorf("could not connect to the leading mesos master")
	default:
		return nil, httpResponseToError(resp)
	}
}

// Roles returns a stripped down mesos/roles containing quota information.
func (c *Client) Roles() (*Roles, error) {
	resp, err := c.http.Get("/roles")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var roles Roles
		err = json.NewDecoder(resp.Body).Decode(&roles)
		return &roles, err
	case 503:
		return nil, fmt.Errorf("could not connect to the leading mesos master")
	default:
		return nil, fmt.Errorf("HTTP %d error", resp.StatusCode)
	}
}

// UpdateQuota updates a quota.
func (c *Client) UpdateQuota(name string, cpu float64, disk float64, gpu float64, mem float64, force bool) error {
	limits := map[string]mesos.Value_Scalar{}
	if cpu >= 0 {
		limits["cpus"] = mesos.Value_Scalar{Value: cpu}
	}
	if disk >= 0 {
		limits["disk"] = mesos.Value_Scalar{Value: disk}
	}
	if gpu >= 0 {
		limits["gpus"] = mesos.Value_Scalar{Value: gpu}
	}
	if mem >= 0 {
		limits["mem"] = mesos.Value_Scalar{Value: mem}
	}

	config := quota.QuotaConfig{
		Role:   name,
		Limits: limits,
	}
	body := master.Call{
		Type: master.Call_UPDATE_QUOTA,
		UpdateQuota: &master.Call_UpdateQuota{
			Force:        &force,
			QuotaConfigs: []quota.QuotaConfig{config},
		},
	}

	var reqBody bytes.Buffer
	if err := json.NewEncoder(&reqBody).Encode(body); err != nil {
		return err
	}
	resp, err := c.http.Post("/api/v1", "application/json", &reqBody, httpclient.FailOnErrStatus(false))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return nil
	case 404:
		return fmt.Errorf("could not set quota '%s'", name)
	default:
		return httpResponseToError(resp)
	}
}

// DeleteQuota deletes a quota.
func (c *Client) DeleteQuota(quota string) error {
	body := master.Call{
		Type: master.Call_REMOVE_QUOTA,
		RemoveQuota: &master.Call_RemoveQuota{
			Role: quota,
		},
	}
	reqBody, err := proto.Marshal(&body)
	if err != nil {
		return err
	}

	resp, err := c.http.Post("/api/v1", "application/x-protobuf", bytes.NewBuffer(reqBody),
		httpclient.Header("Accept", "application/x-protobuf"))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return nil
	case 503:
		return fmt.Errorf("could not connect to the leading mesos master")
	default:
		return httpResponseToError(resp)
	}
}

// TeardownFramework teardowns a framework.
func (c *Client) TeardownFramework(frameworkID string) error {
	body := master.Call{
		Type: master.Call_TEARDOWN,
		Teardown: &master.Call_Teardown{
			FrameworkID: mesos.FrameworkID{Value: frameworkID},
		},
	}
	var reqBody bytes.Buffer
	if err := json.NewEncoder(&reqBody).Encode(body); err != nil {
		return err
	}
	resp, err := c.http.Post("/api/v1", "application/json", &reqBody, httpclient.FailOnErrStatus(false))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return nil
	case 404:
		return fmt.Errorf("could not teardown framework '%s'", frameworkID)
	default:
		return httpResponseToError(resp)
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
