package debug

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dcos/dcos-cli/pkg/config"
	"github.com/dcos/dcos-cli/pkg/mock"
	"github.com/dcos/dcos-core-cli/pkg/marathon"
	"github.com/stretchr/testify/assert"
)

func TestMarathonDebugDetails(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/service/marathon/v2/queue?embed=lastUnusedOffers", r.URL.String())
		assert.Equal(t, http.MethodGet, r.Method)
		w.Write([]byte(`{"queue":[{"count":1,"since":"2019-11-30T17:29:19.156Z","app":{"id":"/app-test"}}]}`))
	}))
	defer ts.Close()

	ctx, out := newContext(ts)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	err = marathonDebugDetails(ctx, client, "/app-test", false)
	assert.NoError(t, err)
	assert.NotEmpty(t, out)
}

func TestMarathonDebugDetailsUnavailableApp(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/service/marathon/v2/queue?embed=lastUnusedOffers", r.URL.String())
		assert.Equal(t, http.MethodGet, r.Method)
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	ctx, out := newContext(ts)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	err = marathonDebugDetails(ctx, client, "/app-test", false)
	assert.EqualError(t, err, "couldn't find app /app-test in Marathon queue")
	assert.Empty(t, out)
}

func newContext(ts *httptest.Server) (*mock.Context, *bytes.Buffer) {
	out := new(bytes.Buffer)
	env := mock.NewEnvironment()
	env.Out = out
	ctx := mock.NewContext(env)
	cluster := config.NewCluster(nil)
	cluster.SetURL(ts.URL)
	ctx.SetCluster(cluster)
	return ctx, out
}
