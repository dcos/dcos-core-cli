package mesos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/dcos/dcos-cli/api"
	"github.com/dcos/dcos-cli/pkg/httpclient"
	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/golang/protobuf/proto"
	mesos "github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/agent"
	"github.com/mesos/mesos-go/api/v1/lib/agent/calls"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli/httpagent"
	"github.com/mesos/mesos-go/api/v1/lib/master"
	"github.com/mesos/mesos-go/api/v1/lib/recordio"
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
		return nil, fmt.Errorf("HTTP %d error", resp.StatusCode)
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
		return nil, fmt.Errorf("HTTP %d error", resp.StatusCode)
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

// LaunchNestedContainer returns the Mesos leader of the connected cluster.
func (c *Client) LaunchNestedContainer(agentID string, containerID mesos.ContainerID) error {
	cmd := "sleep 1000"
	launchNestedContainer := agent.Call_LaunchNestedContainer{
		ContainerID: containerID,
		Command:     &mesos.CommandInfo{Value: &cmd, Arguments: []string{"sleep", "1000"}},
	}
	body := agent.Call{
		Type:                  agent.Call_LAUNCH_NESTED_CONTAINER,
		LaunchNestedContainer: &launchNestedContainer,
	}
	reqBody, err := proto.Marshal(&body)
	if err != nil {
		return err
	}

	// This will only work with DC/OS.
	resp, err := c.http.Post(fmt.Sprintf("/slave/%s/api/v1", agentID), "application/x-protobuf", bytes.NewBuffer(reqBody),
		httpclient.Header("Accept", "application/x-protobuf"))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return nil
	case 503:
		return fmt.Errorf("could not connect to the agent '%s'", agentID)
	default:
		return httpResponseToError(resp)
	}
}

// TaskAttachExec does task attach or exec.
func (c *Client) TaskAttachExec(agentID string, containerID mesos.ContainerID, cmd string, args []string, interactive bool, tty bool) error {
	var (
		ctx, cancel = context.WithCancel(context.TODO())
		// winCh       <-chan mesos.TTYInfo_WindowSize
	)
	if tty {
		ttyd, err := initTTY()
		if err != nil {
			cancel() // stop go-vet from complaining
			return err
		}

		go func() {
			<-ctx.Done()
			//println("closing ttyd via ctx.Done")
			ttyd.Close()
		}()

		// winCh = ttyd.winch
	}

	standardRoundTripper := pluginutil.NewHTTPClient("").Transport
	mesosRoundTripper := httpcli.RoundTripper(standardRoundTripper)
	mesosDoFunc := httpcli.With(mesosRoundTripper)
	mesosCLI := httpcli.New(httpcli.Endpoint(fmt.Sprintf("%s/slave/%s/api/v1", c.http.BaseURL().String(), agentID)), httpcli.Do(mesosDoFunc))
	cli := httpagent.NewSender(mesosCLI.Send)

	// var aciCh = make(chan *agent.Call, 1)            // must be buffered to avoid blocking below
	// aciCh <- calls.AttachContainerInput(containerID) // very first input message MUST be this
	// go func() {
	// 	defer cancel()
	// 	acif := calls.FromChan(aciCh)

	// 	// blocking call, hence the goroutine; Send only returns when the input stream is severed
	// 	err2 := calls.SendNoData(ctx, cli, acif)
	// 	if err2 != nil && err2 != io.EOF {
	// 		log.Printf("attached input stream error %v", err2)
	// 	}

	// 	fmt.Println("Attached input stream worked!")
	// }()

	// Attach to container stdout and stderr. Send returns immediately with a Response from which output may be decoded.
	output, err := cli.Send(ctx, calls.NonStreaming(calls.KillContainer(containerID)))
	if err != nil {
		log.Printf("attach output stream error: %v", err)
		if output != nil {
			output.Close()
		}
		cancel()
		return err
	}

	go func() {
		defer cancel()
		c.AttachContainerOutput(output, os.Stdout, os.Stderr)
	}()

	// go c.AttachContainerInput(ctx, os.Stdin, winCh, aciCh)
	return nil
}

// AttachContainerInput attaches the container input to the terminal.
func (c *Client) AttachContainerInput(ctx context.Context, stdin io.Reader, winCh <-chan mesos.TTYInfo_WindowSize, aciCh chan<- *agent.Call) {
	defer close(aciCh)

	input := make(chan []byte)
	go func() {
		defer close(input)
		escape := []byte{0x10, 0x11} // CTRL-P, CTRL-Q
		var last byte
		for {
			buf := make([]byte, 512)
			n, err := stdin.Read(buf)
			if n > 0 {
				if (last == escape[0] && buf[0] == escape[1]) || bytes.Index(buf, escape) > -1 {
					return
				}
				buf = buf[:n]
				last = buf[n-1]
				select {
				case input <- buf:
				case <-ctx.Done():
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-input:
			if !ok {
				return
			}
			c := calls.AttachContainerInputData(data)
			select {
			case aciCh <- c:
			case <-ctx.Done():
				return
			}
		case ws := <-winCh:
			c := calls.AttachContainerInputTTY(&mesos.TTYInfo{WindowSize: &ws})
			select {
			case aciCh <- c:
			case <-ctx.Done():
				return
			}
		}
	}
}

// NewbieAttachContainerOutput does stuff
func (c *Client) NewbieAttachContainerOutput(agentID string, containerID mesos.ContainerID) error {
	fmt.Println("Yolo " + containerID.GetValue())
	body := agent.Call{
		Type: agent.Call_ATTACH_CONTAINER_OUTPUT,
		AttachContainerOutput: &agent.Call_AttachContainerOutput{
			ContainerID: containerID,
		},
	}
	reqBody, err := json.Marshal(&body)
	if err != nil {
		return err
	}

	resp, err := c.http.Post(fmt.Sprintf("/slave/%s/api/v1", agentID), "application/json", bytes.NewBuffer(reqBody), httpclient.Header("Accept", "application/recordio"), httpclient.Header("Message-Accept", "application/json"))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		// io.Copy(os.Stdout, resp.Body)
		r := recordio.NewReader(resp.Body)
		for {
			data, err := r.ReadFrame()
			if err != nil {
				return err
			}
			var pio agent.ProcessIO
			json.Unmarshal(data, &pio)
			switch pio.GetType() {
			case agent.ProcessIO_DATA:
				data := pio.GetData()
				switch data.GetType() {
				case agent.ProcessIO_Data_STDOUT:
					os.Stdout.Write(data.GetData())
				case agent.ProcessIO_Data_STDERR:
					os.Stderr.Write(data.GetData())
				default:
				}
			}
		}
		/*
			mesosEncodingProto.NewDecoder(mesosEncoding.NewSource(resp.Body))
			var pio agent.ProcessIO
			err := resp.Decode(&pio)
			if err != nil {
				return err
			}
			switch pio.GetType() {
			case agent.ProcessIO_DATA:
				data := pio.GetData()
				switch data.GetType() {
				case agent.ProcessIO_Data_STDOUT:
					os.Stdout.Write(data.GetData())
				case agent.ProcessIO_Data_STDERR:
					os.Stderr.Write(data.GetData())
				default:
				}
			default:
			}
		*/
		//}
		return nil
	case 503:
		return fmt.Errorf("could not connect to the leading mesos master")
	default:
		return httpResponseToError(resp)
	}
}

// AttachContainerOutput attaches the container output to the terminal.
func (c *Client) AttachContainerOutput(resp mesos.Response, stdout, stderr io.Writer) error {
	defer resp.Close()
	forward := func(b []byte, out io.Writer) error {
		n, err := out.Write(b)
		if err == nil && len(b) != n {
			err = io.ErrShortWrite
		}
		return err
	}
	for {
		var pio agent.ProcessIO
		err := resp.Decode(&pio)
		if err != nil {
			return err
		}
		switch pio.GetType() {
		case agent.ProcessIO_DATA:
			data := pio.GetData()
			switch data.GetType() {
			case agent.ProcessIO_Data_STDOUT:
				if err := forward(data.GetData(), stdout); err != nil {
					return err
				}
			case agent.ProcessIO_Data_STDERR:
				if err := forward(data.GetData(), stderr); err != nil {
					return err
				}
			default:
			}
		default:
		}
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
