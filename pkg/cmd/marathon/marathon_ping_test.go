package marathon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dcos/dcos-core-cli/pkg/marathon"
	"github.com/dcos/dcos-core-cli/pkg/mesos"
	"github.com/stretchr/testify/assert"
)

func TestMarathonPingOnce(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/service/marathon/ping", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		w.Write([]byte(`pong`))
	}))
	defer ts.Close()

	ctx := newContext(ts)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	masters, err := marathonPing(ctx, *client, true)
	assert.Equal(t, 1, masters)
	assert.NoError(t, err)
}

func TestMarathonPing(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch url := r.URL.String(); url {
		case "/service/marathon/ping":
			w.Write([]byte(`pong`))
		case "/mesos_dns/v1/hosts/master.mesos":
			masters := []mesos.Master{{}, {}}
			data, err := json.Marshal(masters)
			assert.NoError(t, err)
			w.Write(data)
		}
	}))
	defer ts.Close()

	ctx := newContext(ts)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	masters, err := marathonPing(ctx, *client, false)
	assert.Equal(t, 2, masters)
	assert.NoError(t, err)
}

func TestMarathonPingError(t *testing.T) {
	pingPinged := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch url := r.URL.String(); url {
		case "/service/marathon/ping":
			if pingPinged == 0 {
				w.Write([]byte(`pong`))
				pingPinged++
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("500 - Something bad happened!"))
			}
		case "/mesos_dns/v1/hosts/master.mesos":
			masters := []mesos.Master{{}, {}}
			data, err := json.Marshal(masters)
			assert.NoError(t, err)
			w.Write(data)
		}
	}))
	defer ts.Close()

	ctx := newContext(ts)

	client, err := marathon.NewClient(ctx)
	assert.NoError(t, err)

	masters, err := marathonPing(ctx, *client, false)
	assert.Equal(t, 0, masters)
	assert.EqualError(t, err, `unable to ping leading Marathon master 2 time(s)`)
}
