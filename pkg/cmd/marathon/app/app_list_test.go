package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dcos/dcos-cli/pkg/config"
	"github.com/dcos/dcos-cli/pkg/mock"
	"github.com/dcos/dcos-core-cli/pkg/marathon"
	m "github.com/gambol99/go-marathon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppList(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch url := r.URL.String(); url {
		case "/service/marathon/v2/apps":
			apps := m.Applications{
				Apps: []m.Application{
					{
						ID:           "test-id",
						CPUs:         1.0,
						Mem:          floatPointer(2.0),
						Cmd:          strPointer("test cmd"),
						Role:         strPointer("public_agent"),
						TasksRunning: 3,
						Instances:    intPointer(4),
					},
				},
			}
			json.NewEncoder(w).Encode(apps)
		case "/service/marathon/v2/deployments":
			deployments := []m.Deployment{}
			json.NewEncoder(w).Encode(deployments)
		case "/service/marathon/v2/queue":
			queue := m.Queue{}
			json.NewEncoder(w).Encode(queue)
		}
	}))
	defer ts.Close()

	out := new(bytes.Buffer)
	env := mock.NewEnvironment()
	env.Out = out
	ctx := mock.NewContext(env)
	cluster := config.NewCluster(nil)
	cluster.SetURL(ts.URL)
	ctx.SetCluster(cluster)

	client, err := marathon.NewClient(ctx)
	require.NoError(t, err)

	err = appList(ctx, client, false, false)
	assert.NoError(t, err)
	expected := "    ID     MEM  CPUS  TASKS  HEALTH  DEPLOYMENT  WAITING  CONTAINER    CMD         ROLE      \n" +
		"  test-id  2    1     3/4    ---     ---         false    MESOS      test cmd  public_agent  \n"
	assert.Equal(t, expected, out.String())
}

func TestAppListQuiet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apps := m.Applications{
			Apps: []m.Application{
				{
					ID:           "test-id",
					CPUs:         1.0,
					Mem:          floatPointer(2.0),
					Cmd:          strPointer("test cmd"),
					Role:         strPointer("public_agent"),
					TasksRunning: 3,
					Instances:    intPointer(4),
				},
			},
		}
		json.NewEncoder(w).Encode(apps)
	}))
	defer ts.Close()

	out := new(bytes.Buffer)
	env := mock.NewEnvironment()
	env.Out = out
	ctx := mock.NewContext(env)
	cluster := config.NewCluster(nil)
	cluster.SetURL(ts.URL)
	ctx.SetCluster(cluster)

	client, err := marathon.NewClient(ctx)
	require.NoError(t, err)

	err = appList(ctx, client, true, false)
	assert.NoError(t, err)
	assert.Equal(t, "test-id\n", out.String())

}

func TestAppListJSON(t *testing.T) {
	apps := m.Applications{
		Apps: []m.Application{
			{
				ID:           "test-id",
				CPUs:         1.0,
				Mem:          floatPointer(2.0),
				Cmd:          strPointer("test cmd"),
				Role:         strPointer("public_agent"),
				TasksRunning: 3,
				Instances:    intPointer(4),
			},
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(apps)
	}))
	defer ts.Close()

	out := new(bytes.Buffer)
	env := mock.NewEnvironment()
	env.Out = out
	ctx := mock.NewContext(env)
	cluster := config.NewCluster(nil)
	cluster.SetURL(ts.URL)
	ctx.SetCluster(cluster)

	client, err := marathon.NewClient(ctx)
	require.NoError(t, err)

	err = appList(ctx, client, false, true)
	assert.NoError(t, err)

	expected := new(bytes.Buffer)
	enc := json.NewEncoder(expected)
	enc.SetIndent("", "    ")
	err = enc.Encode(apps.Apps)
	require.NoError(t, err)

	assert.Equal(t, expected.String(), out.String())
}

// floatPointer returns a pointer to given float, helpful since Go doesn't
// allow declarations of a constant float pointer
func floatPointer(f float64) *float64 {
	return &f
}

func intPointer(i int) *int {
	return &i
}

func strPointer(s string) *string {
	return &s
}
