package marathon

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dcos/dcos-cli/pkg/config"
	"github.com/dcos/dcos-cli/pkg/mock"
	"github.com/dcos/dcos-core-cli/pkg/marathon"
	"github.com/stretchr/testify/assert"
)

func TestMarathonAbout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/service/marathon/v2/info", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		w.Write([]byte(`{"frameworkId":"marathonId"}`))
	}))
	defer ts.Close()

	env := mock.NewEnvironment()
	ctx := mock.NewContext(env)
	cluster := config.NewCluster(nil)
	cluster.SetURL(ts.URL)
	ctx.SetCluster(cluster)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	err = marathonAbout(ctx, *client)
	assert.NoError(t, err)
}

func TestMarathonAboutClientError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/service/marathon/v2/info", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Something bad happened!"))
	}))
	defer ts.Close()

	env := mock.NewEnvironment()
	ctx := mock.NewContext(env)
	cluster := config.NewCluster(nil)
	cluster.SetURL(ts.URL)
	ctx.SetCluster(cluster)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	err = marathonAbout(ctx, *client)
	assert.EqualError(t, err, `HTTP 500 error`)
}
