package mesos

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dcos/dcos-core-cli/pkg/pluginutil"
	"github.com/golang/protobuf/proto"
	mesos "github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/master"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHosts(t *testing.T) {
	expectedHosts := []Host{
		Host{
			Host: "dcos.example.org",
			IP:   "8.8.8.8",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/mesos_dns/v1/hosts/8.8.8.8", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		assert.NoError(t, json.NewEncoder(w).Encode(expectedHosts))
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL))

	hosts, err := c.Hosts("8.8.8.8")
	require.NoError(t, err)
	require.Equal(t, expectedHosts, hosts)
}

func TestLeader(t *testing.T) {
	expectedHosts := []Master{
		Master{
			Host: "leader.mesos",
			IP:   "10.0.0.10",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/mesos_dns/v1/hosts/leader.mesos", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		assert.NoError(t, json.NewEncoder(w).Encode(expectedHosts))
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL))

	leader, err := c.Leader()
	require.NoError(t, err)
	require.Equal(t, &expectedHosts[0], leader)
}

func TestMasters(t *testing.T) {
	expectedHosts := []Master{
		Master{
			Host: "master.mesos",
			IP:   "10.0.0.10",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/mesos_dns/v1/hosts/master.mesos", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		assert.NoError(t, json.NewEncoder(w).Encode(expectedHosts))
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL))

	masters, err := c.Masters()
	require.NoError(t, err)
	require.Equal(t, expectedHosts, masters)
}

func TestState(t *testing.T) {
	expectedState := State{
		ID:       "2a2a4995-3a44-41b8-9610-1e0bca965e88",
		PID:      "10.0.0.10@2a2a4995-3a44-41b8-9610-1e0bca965e88",
		Hostname: "master.mesos",
		Cluster:  "cluster-ops-cli-test-cluster-fcdb4f27",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/master/state", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		assert.NoError(t, json.NewEncoder(w).Encode(expectedState))
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL))

	s, err := c.State()
	require.NoError(t, err)
	require.Equal(t, &expectedState, s)
}

func TestStateSummary(t *testing.T) {
	expectedState := State{
		Hostname: "master.mesos",
		Cluster:  "cluster-ops-cli-test-cluster-fcdb4f27",
		Slaves: []Slave{
			Slave{
				ID:       "2a2a4995-3a44-41b8-9610-1e0bca965e88-S0",
				PID:      "10.0.0.23@2a2a4995-3a44-41b8-9610-1e0bca965e88-S0",
				Hostname: "10.0.0.23",
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/master/state-summary", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		assert.NoError(t, json.NewEncoder(w).Encode(expectedState))
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL))

	ss, err := c.StateSummary()
	require.NoError(t, err)
	require.Equal(t, expectedState.Hostname, ss.Hostname)
	require.Equal(t, expectedState.Cluster, ss.Cluster)
	require.Equal(t, expectedState.Slaves, ss.Slaves)
}

func TestAgents(t *testing.T) {
	expectedAgents := master.Response{
		GetAgents: &master.Response_GetAgents{
			Agents: []master.Response_GetAgents_Agent{
				master.Response_GetAgents_Agent{
					Active: true,
					AgentInfo: mesos.AgentInfo{
						Hostname: "10.0.0.23",
					},
				},
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1", r.URL.String())
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/x-protobuf", r.Header.Get("Accept"))
		assert.Equal(t, "application/x-protobuf", r.Header.Get("Content-Type"))
		response, err := proto.Marshal(&expectedAgents)
		assert.NoError(t, err)
		w.Write(response)
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL))

	agents, err := c.Agents()
	require.NoError(t, err)
	require.Equal(t, expectedAgents.GetAgents.Agents, agents)
}

func TestMarkAgentGone(t *testing.T) {
	const expectedAgentID = "9001"
	expectedBody := master.Call{
		Type: master.Call_MARK_AGENT_GONE,
		MarkAgentGone: &master.Call_MarkAgentGone{
			AgentID: mesos.AgentID{Value: expectedAgentID},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1", r.URL.String())
		assert.Equal(t, "POST", r.Method)
		var payload master.Call
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)
		assert.Equal(t, expectedBody, payload)
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL))

	err := c.MarkAgentGone(expectedAgentID)
	require.NoError(t, err)
}

func TestTeardownFramework(t *testing.T) {
	const expectedFrameworkID = "yolo"
	expectedBody := master.Call{
		Type: master.Call_TEARDOWN,
		Teardown: &master.Call_Teardown{
			FrameworkID: mesos.FrameworkID{Value: expectedFrameworkID},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1", r.URL.String())
		assert.Equal(t, "POST", r.Method)
		var payload master.Call
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)
		assert.Equal(t, expectedBody, payload)
	}))
	defer ts.Close()

	c := NewClient(pluginutil.HTTPClient(ts.URL))

	err := c.TeardownFramework(expectedFrameworkID)
	require.NoError(t, err)
}
