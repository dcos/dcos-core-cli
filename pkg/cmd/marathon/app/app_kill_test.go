package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dcos/dcos-core-cli/pkg/marathon"
	goMarathon "github.com/gambol99/go-marathon"
	"github.com/stretchr/testify/assert"
)

func TestMarathonKill(t *testing.T) {
	tasks := `[{"appId":"/test","healthCheckResults":[{"alive":true,"consecutiveFailures":0,"instanceId":"id"}]}]`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/service/marathon/v2/apps/app-test/tasks", r.URL.String())
		assert.Equal(t, "DELETE", r.Method)
		w.Write([]byte("{\"tasks\":" + tasks + "}"))
	}))
	defer ts.Close()

	ctx, out := newContext(ts)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	err = marathonAppKill(ctx, *client, "app-test", false, "")
	assert.NoError(t, err)

	assert.Equal(t, fmt.Sprintf("Killed tasks: %s\n", tasks), out.String())
}

func TestMarathonKillHost(t *testing.T) {
	tasks := `[{"appId":"/test","healthCheckResults":[{"alive":true,"consecutiveFailures":0,"instanceId":"id"}]}]`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/service/marathon/v2/apps/app-test/tasks?host=192.168.0.1", r.URL.String())
		assert.Equal(t, "DELETE", r.Method)
		w.Write([]byte("{\"tasks\":" + tasks + "}"))
	}))
	defer ts.Close()

	ctx, out := newContext(ts)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	err = marathonAppKill(ctx, *client, "app-test", false, "192.168.0.1")
	assert.NoError(t, err)

	assert.Equal(t, fmt.Sprintf("Killed tasks: %s\n", tasks), out.String())
}

func TestMarathonKillMissingApp(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/service/marathon/v2/apps/app-test/tasks", r.URL.String())
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Something bad happened!"))
	}))
	defer ts.Close()

	ctx, _ := newContext(ts)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	err = marathonAppKill(ctx, *client, "app-test", false, "")
	assert.EqualError(t, err, `app '/app-test' does not exist`)
}

func TestMarathonKillScale(t *testing.T) {
	deployment := goMarathon.DeploymentID{
		DeploymentID: "29147919-b4f0-47d0-93d6-2ed9a2be1d8c",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/service/marathon/v2/apps/app-test", r.URL.String())
		switch method := r.Method; method {
		case "GET":
			w.Write([]byte(`{"apps": [{"id": "/app-test"}]}`))
		case "PUT":
			w.Header().Set("Marathon-Deployment-Id", deployment.DeploymentID)
		}
	}))
	defer ts.Close()

	ctx, out := newContext(ts)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	err = marathonAppKill(ctx, *client, "app-test", true, "")
	assert.NoError(t, err)

	deploymentString, err := json.Marshal(deployment)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("Started deployment: %s\n", deploymentString), out.String())
}

func TestMarathonKillScaleMissingApp(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/service/marathon/v2/apps/app-test", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Something bad happened!"))
	}))
	defer ts.Close()

	ctx, _ := newContext(ts)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	err = marathonAppKill(ctx, *client, "app-test", true, "")
	assert.EqualError(t, err, `path '/app-test' does not exist`)
}
