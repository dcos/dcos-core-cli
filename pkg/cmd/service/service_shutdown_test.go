package service

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/dcos/dcos-cli/pkg/config"
	"github.com/dcos/dcos-cli/pkg/mock"
)

func Test_serviceShutdown(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		yes         bool
		input       string
		output      string
		err         error
	}{
		{name: "empty service name", yes: true, err: fmt.Errorf("service name must not be empty")},
		{name: "no confirmation", serviceName: "test-service", err: fmt.Errorf("couldn't get confirmation"),
			output: "Do you really want to teardown test-service with all its tasks? [yes/no] "},
		{name: "no error", serviceName: "test-service", yes: true},
		{name: "no error with confirmation", serviceName: "test-service", input: "Y\n",
			output: "Do you really want to teardown test-service with all its tasks? [yes/no] "},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, out := newContext(ts, tt.input)
			err := serviceShutdown(ctx, tt.serviceName, tt.yes)
			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.output, out.String())
		})
	}
}

func Test_serviceShutdownWithoutCluster(t *testing.T) {
	ctx := mock.NewContext(nil)
	err := serviceShutdown(ctx, "test-service", true)
	assert.EqualError(t, err, "no cluster is attached")

}

func newContext(ts *httptest.Server, input string) (*mock.Context, *bytes.Buffer) {
	out := new(bytes.Buffer)
	env := mock.NewEnvironment()
	env.Input = strings.NewReader(input)
	env.Out = out
	env.ErrOut = out
	ctx := mock.NewContext(env)
	// Use tempdir as cluster dir to check if path if properly created
	conf := config.Empty()
	conf.SetPath(path.Join(os.TempDir(), "conf.toml"))
	cluster := config.NewCluster(conf)
	cluster.SetURL(ts.URL)
	ctx.SetCluster(cluster)
	ctx.Logger().Out = out
	ctx.Logger().SetLevel(logrus.DebugLevel)
	ctx.Logger().SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	return ctx, out
}
