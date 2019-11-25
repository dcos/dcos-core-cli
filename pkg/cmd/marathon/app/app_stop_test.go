package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dcos/dcos-cli/pkg/config"
	"github.com/dcos/dcos-cli/pkg/mock"
	goMarathon "github.com/gambol99/go-marathon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppStop(t *testing.T) {
	app := struct {
		App goMarathon.Application `json:"app"`
	}{
		App: goMarathon.Application{
			ID:        "test",
			Instances: intPointer(3),
		},
	}
	deployment := goMarathon.DeploymentID{
		DeploymentID: "5ed4c0c5-9ff8-4a6f-a0cd-f57f59a34b43",
		Version:      "2015-09-29T15:59:51.164Z",
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch url := r.URL.String(); url {
		//nolint:goconst
		case "/service/marathon/v2/apps/test":
			switch r.Method {
			case http.MethodGet:
				w.WriteHeader(200)
				json.NewEncoder(w).Encode(app)
			case http.MethodPut:
				w.WriteHeader(200)
				json.NewEncoder(w).Encode(deployment)
			default:
				require.Fail(t, "unexpected http verb", r.Method)
			}
		default:
			require.Fail(t, "unexpected call to endpoint", url)
		}
	}))

	out := new(bytes.Buffer)
	env := mock.NewEnvironment()
	env.Out = out
	ctx := mock.NewContext(env)
	cluster := config.NewCluster(nil)
	cluster.SetURL(ts.URL)
	ctx.SetCluster(cluster)

	err := appStop(ctx, "/test", false)
	require.NoError(t, err)

	assert.Equal(t, "Created deployment 5ed4c0c5-9ff8-4a6f-a0cd-f57f59a34b43\n", out.String())

}

func TestAppStopWhenAppIsStopped(t *testing.T) {
	app := struct {
		App goMarathon.Application `json:"app"`
	}{
		App: goMarathon.Application{
			ID:        "test",
			Instances: intPointer(0),
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch url := r.URL.String(); url {
		case "/service/marathon/v2/apps/test":
			switch r.Method {
			case http.MethodGet:
				w.WriteHeader(200)
				json.NewEncoder(w).Encode(app)
			default:
				require.Fail(t, "unexpected http verb", url, r.Method)
			}
		default:
			require.Fail(t, "unexpected call to endpoint", url)
		}
	}))

	out := new(bytes.Buffer)
	env := mock.NewEnvironment()
	env.Out = out
	ctx := mock.NewContext(env)
	cluster := config.NewCluster(nil)
	cluster.SetURL(ts.URL)
	ctx.SetCluster(cluster)

	err := appStop(ctx, "/test", false)
	assert.EqualError(t, err, "app '/test' already stopped: 0 instances")
}

func TestAppStopWhenAppDoesntExist(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch url := r.URL.String(); url {
		case "/service/marathon/v2/apps/test":
			w.WriteHeader(http.StatusNotFound)
		default:
			require.Fail(t, "unexpected call to endpoint", url)
		}
	}))

	out := new(bytes.Buffer)
	env := mock.NewEnvironment()
	env.Out = out
	ctx := mock.NewContext(env)
	cluster := config.NewCluster(nil)
	cluster.SetURL(ts.URL)
	ctx.SetCluster(cluster)
	err := appStop(ctx, "/test", false)
	assert.EqualError(t, err, "app '/test' does not exist")
}

func TestAppStopForce(t *testing.T) {
	app := struct {
		App goMarathon.Application `json:"app"`
	}{
		App: goMarathon.Application{
			ID:        "test",
			Instances: intPointer(2),
		},
	}
	deployment := goMarathon.DeploymentID{
		DeploymentID: "5ed4c0c5-9ff8-4a6f-a0cd-f57f59a34b43",
		Version:      "2015-09-29T15:59:51.164Z",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch url := r.URL.Path; url {
		case "/service/marathon/v2/apps/test":
			switch r.Method {
			case http.MethodGet:
				w.WriteHeader(200)
				json.NewEncoder(w).Encode(app)
			case http.MethodPut:
				assert.Contains(t, r.URL.Query(), "force")
				w.WriteHeader(200)
				json.NewEncoder(w).Encode(deployment)
			default:
				require.Fail(t, "unexpected http verb", url, r.Method)
			}
		default:
			require.Fail(t, "unexpected call to endpoint", url)
		}
	}))

	out := new(bytes.Buffer)
	env := mock.NewEnvironment()
	env.Out = out
	ctx := mock.NewContext(env)
	cluster := config.NewCluster(nil)
	cluster.SetURL(ts.URL)
	ctx.SetCluster(cluster)

	err := appStop(ctx, "/test", true)
	assert.NoError(t, err)
}

func TestAppStopWhenDeploymentBlocked(t *testing.T) {
	app := struct {
		App goMarathon.Application `json:"app"`
	}{
		App: goMarathon.Application{
			ID:        "test",
			Instances: intPointer(2),
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch url := r.URL.String(); url {
		case "/service/marathon/v2/apps/test":
			switch r.Method {
			case http.MethodGet:
				w.WriteHeader(200)
				json.NewEncoder(w).Encode(app)
			case http.MethodPut:
				w.WriteHeader(409)
			default:
				require.Fail(t, "unexpected http verb", url, r.Method)
			}
		default:
			require.Fail(t, "unexpected call to endpoint", url)
		}
	}))

	out := new(bytes.Buffer)
	env := mock.NewEnvironment()
	env.Out = out
	ctx := mock.NewContext(env)
	cluster := config.NewCluster(nil)
	cluster.SetURL(ts.URL)
	ctx.SetCluster(cluster)

	err := appStop(ctx, "/test", false)
	assert.EqualError(t, err, "changes blocked: deployment already in progress for app")
}
