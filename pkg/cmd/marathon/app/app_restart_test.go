package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dcos/dcos-core-cli/pkg/marathon"
	"github.com/stretchr/testify/assert"
)

func TestMarathonRestart(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch method := r.Method; method {
		case http.MethodGet:
			assert.Equal(t, "/service/marathon/v2/apps/app-test", r.URL.String())
			w.Write([]byte(`{"app":{"id":"/app-test","backoffFactor":1.15,"backoffSeconds":1,"cpus":1.0,"disk":64.0,"executor":"","instances":1}}`))
		case http.MethodPost:
			assert.Equal(t, "/service/marathon/v2/apps/app-test/restart", r.URL.String())
			w.Write([]byte(`{"deploymentId":"5ed4c0c5-9ff8-4a6f-a0cd-f57f59a34b43","version":"2015-09-29T15:59:51.164Z"}`))
		}
	}))
	defer ts.Close()

	ctx, _ := newContext(ts)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	deployment, err := marathonAppRestart(*client, "app-test", false)
	assert.Equal(t, "5ed4c0c5-9ff8-4a6f-a0cd-f57f59a34b43", deployment.DeploymentID)
	assert.NoError(t, err)
}

func TestMarathonRestartForce(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch method := r.Method; method {
		case http.MethodGet:
			assert.Equal(t, "/service/marathon/v2/apps/app-test", r.URL.String())
			w.Write([]byte(`{"app":{"id":"/app-test","backoffFactor":1.15,"backoffSeconds":1,"cpus":1.0,"disk":64.0,"executor":"","instances":1}}`))
		case http.MethodPost:
			assert.Equal(t, "/service/marathon/v2/apps/app-test/restart?force=true", r.URL.String())
			w.Write([]byte(`{"deploymentId":"5ed4c0c5-9ff8-4a6f-a0cd-f57f59a34b43","version":"2015-09-29T15:59:51.164Z"}`))
		}
	}))
	defer ts.Close()

	ctx, _ := newContext(ts)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	deployment, err := marathonAppRestart(*client, "app-test", true)
	assert.Equal(t, "5ed4c0c5-9ff8-4a6f-a0cd-f57f59a34b43", deployment.DeploymentID)
	assert.NoError(t, err)
}

func TestMarathonRestartMissingApp(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/service/marathon/v2/apps/app-test", r.URL.String())
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("App '/app-test' does not exist"))
	}))
	defer ts.Close()

	ctx, _ := newContext(ts)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	_, err = marathonAppRestart(*client, "app-test", false)
	assert.EqualError(t, err, `app '/app-test' does not exist`)
}

func TestMarathonRestartNoRunningTasks(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/service/marathon/v2/apps/app-test", r.URL.String())
		w.Write([]byte(`{"app":{"id":"/app-test","backoffFactor":1.15,"backoffSeconds":1,"cpus":1.0,"disk":64.0,"executor":"","instances":0}}`))
	}))
	defer ts.Close()

	ctx, _ := newContext(ts)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	_, err = marathonAppRestart(*client, "app-test", false)
	assert.EqualError(t, err, `unable to perform rolling restart of application '/app-test' because it has no running tasks`)
}
