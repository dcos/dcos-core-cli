package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dcos/dcos-core-cli/pkg/marathon"
	"github.com/stretchr/testify/assert"
)

func TestMarathonStart(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/service/marathon/v2/apps/app-test", r.URL.String())
		switch method := r.Method; method {
		case http.MethodGet:
			w.Write([]byte(`{"app":{"id":"/app-test","backoffFactor":1.15,"backoffSeconds":1,"cpus":1.0,"disk":64.0,"executor":"","instances":0}}`))
		case http.MethodPut:
			w.Write([]byte(`{"deploymentId":"5ed4c0c5-9ff8-4a6f-a0cd-f57f59a34b43","version":"2015-09-29T15:59:51.164Z"}`))
		}
	}))
	defer ts.Close()

	ctx, _ := newContext(ts)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	deployment, err := marathonAppStart(*client, "app-test", 1, false)
	assert.Equal(t, "5ed4c0c5-9ff8-4a6f-a0cd-f57f59a34b43", deployment.DeploymentID)
	assert.NoError(t, err)
}

func TestMarathonStartNotEnoughInstances(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer ts.Close()

	ctx, _ := newContext(ts)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	_, err = marathonAppStart(*client, "app-test", 0, false)
	assert.EqualError(t, err, "the number of instances must be positive: 0")
}

func TestMarathonStartTooManyInstances(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/service/marathon/v2/apps/app-test", r.URL.String())
		w.Write([]byte(`{"app":{"id":"/app-test","backoffFactor":1.15,"backoffSeconds":1,"cpus":1.0,"disk":64.0,"executor":"","instances":3}}`))
	}))
	defer ts.Close()

	ctx, _ := newContext(ts)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	_, err = marathonAppStart(*client, "app-test", 1, false)
	assert.EqualError(t, err, "application 'app-test' already started: 3 instances")
}

func TestMarathonStartError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/service/marathon/v2/apps/app-test", r.URL.String())
		switch method := r.Method; method {
		case http.MethodGet:
			w.Write([]byte(`{"app":{"id":"/app-test","backoffFactor":1.15,"backoffSeconds":1,"cpus":1.0,"disk":64.0,"executor":"","instances":0}}`))
		case http.MethodPut:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 - Something bad happened!"))
		}
	}))
	defer ts.Close()

	ctx, _ := newContext(ts)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	_, err = marathonAppStart(*client, "app-test", 1, false)
	assert.EqualError(t, err, "Marathon API error: 404 - Something bad happened!")
}

func TestMarathonStartAppDoesNotExist(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/service/marathon/v2/apps/app-test", r.URL.String())
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Something bad happened!"))
	}))
	defer ts.Close()

	ctx, _ := newContext(ts)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	_, err = marathonAppStart(*client, "app-test", 1, false)
	assert.EqualError(t, err, "app '/app-test' does not exist")
}
