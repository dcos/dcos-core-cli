package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dcos/dcos-cli/pkg/config"
	"github.com/dcos/dcos-cli/pkg/mock"
	"github.com/dcos/dcos-core-cli/pkg/marathon"
	"github.com/stretchr/testify/assert"
)

func TestMarathonRemove(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/service/marathon/v2/apps/app-test", r.URL.String())
		assert.Equal(t, "DELETE", r.Method)
		w.Write([]byte(`{"deploymentId":"5ed4c0c5-9ff8-4a6f-a0cd-f57f59a34b43","version":"2015-09-29T15:59:51.164Z"}`))
	}))
	defer ts.Close()

	env := mock.NewEnvironment()
	ctx := mock.NewContext(env)
	cluster := config.NewCluster(nil)
	cluster.SetURL(ts.URL)
	ctx.SetCluster(cluster)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	err = marathonAppRemove(*client, "app-test", false)
	assert.NoError(t, err)
}

func TestMarathonRemoveForce(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/service/marathon/v2/apps/app-test?force=true", r.URL.String())
		assert.Equal(t, "DELETE", r.Method)
		w.Write([]byte(`{"deploymentId":"5ed4c0c5-9ff8-4a6f-a0cd-f57f59a34b43","version":"2015-09-29T15:59:51.164Z"}`))
	}))
	defer ts.Close()

	env := mock.NewEnvironment()
	ctx := mock.NewContext(env)
	cluster := config.NewCluster(nil)
	cluster.SetURL(ts.URL)
	ctx.SetCluster(cluster)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	err = marathonAppRemove(*client, "app-test", true)
	assert.NoError(t, err)
}

func TestMarathonRemoveMissingApp(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/service/marathon/v2/apps/app-test", r.URL.String())
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	env := mock.NewEnvironment()
	ctx := mock.NewContext(env)
	cluster := config.NewCluster(nil)
	cluster.SetURL(ts.URL)
	ctx.SetCluster(cluster)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	err = marathonAppRemove(*client, "app-test", false)
	assert.EqualError(t, err, `app '/app-test' does not exist`)
}

func TestMarathonRemoveOtherError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/service/marathon/v2/apps/app-test", r.URL.String())
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusExpectationFailed)
		w.Write([]byte("Something bad happened"))
	}))
	defer ts.Close()

	env := mock.NewEnvironment()
	ctx := mock.NewContext(env)
	cluster := config.NewCluster(nil)
	cluster.SetURL(ts.URL)
	ctx.SetCluster(cluster)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	err = marathonAppRemove(*client, "app-test", false)
	assert.EqualError(t, err, "Marathon API error: Something bad happened")
}
