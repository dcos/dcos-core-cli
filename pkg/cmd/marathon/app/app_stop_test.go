package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dcos/dcos-cli/pkg/config"
	"github.com/dcos/dcos-cli/pkg/mock"
	goMarathon "github.com/gambol99/go-marathon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testID = "test"

type appResponse struct {
	App goMarathon.Application `json:"app"`
}

func TestAppStopTable(t *testing.T) {
	simpleApp := appResponse{
		App: goMarathon.Application{
			ID:        testID,
			Instances: intPointer(2),
		},
	}

	tests := []struct {
		name    string
		force   bool
		getFunc http.HandlerFunc
		putFunc http.HandlerFunc
		out     string
		err     error
	}{
		{
			getFunc: encodeFunc(simpleApp),
			putFunc: encodeFunc(goMarathon.DeploymentID{
				DeploymentID: "5ed4c0c5-9ff8-4a6f-a0cd-f57f59a34b43",
				Version:      "2015-09-29T15:59:51.164Z",
			}),
			out: "Created deployment 5ed4c0c5-9ff8-4a6f-a0cd-f57f59a34b43\n",
		},
		{
			name: "when app is stopped",
			getFunc: encodeFunc(struct {
				App goMarathon.Application `json:"app"`
			}{App: goMarathon.Application{
				ID:        testID,
				Instances: intPointer(0),
			}}),
			err: errors.New("app '/test' already stopped: 0 instances"),
		},
		{
			name:    "when app doesn't exist",
			getFunc: http.NotFound,
			err:     errors.New("app '/test' does not exist"),
		},
		{
			name:    "with force",
			force:   true,
			getFunc: encodeFunc(simpleApp),
			putFunc: func(w http.ResponseWriter, r *http.Request) {
				assert.Contains(t, r.URL.Query(), "force")
				encode(w, goMarathon.DeploymentID{
					DeploymentID: "5ed4c0c5-9ff8-4a6f-a0cd-f57f59a34b43",
					Version:      "2015-09-29T15:59:51.164Z",
				})
			},
			out: "Created deployment 5ed4c0c5-9ff8-4a6f-a0cd-f57f59a34b43\n",
		},
		{
			name:    "when deployment is blocked",
			getFunc: encodeFunc(simpleApp),
			putFunc: func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(409) },
			err:     errors.New("changes blocked: deployment already in progress for app"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := router(tt.getFunc, tt.putFunc)
			ctx, out := newContext(ts)

			err := appStop(ctx, "/test", tt.force)

			require.Equal(t, tt.err, err)
			assert.Equal(t, tt.out, out.String())
		})
	}
}

func encodeFunc(obj interface{}) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) { encode(w, obj) }
}

// nolint: interfacer
func encode(w http.ResponseWriter, obj interface{}) {
	json.NewEncoder(w).Encode(obj)
}

func router(getFunc http.HandlerFunc, putFunc http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch url := r.URL.Path; url {
		// nolint: goconst
		case "/service/marathon/v2/apps/test":
			switch r.Method {
			case http.MethodGet:
				getFunc(w, r)
			case http.MethodPut:
				putFunc(w, r)
			default:
				panic("unexpected http verb " + r.Method)
			}
		default:
			panic("unexpected call to endpoint " + url)
		}
	}))
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
