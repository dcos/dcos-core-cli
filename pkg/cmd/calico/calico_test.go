package calico

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/dcos/dcos-cli/pkg/config"
	"github.com/dcos/dcos-cli/pkg/mock"
	"gotest.tools/assert"
)

func Test_getMesosState(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/mesos/master/state", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		w.Write([]byte(`{"Hostname":"127.0.0.1"}`))
	}))
	defer ts.Close()
	ctx := newContext(ts)

	data := <-getMesosState(ctx)
	assert.Equal(t, data.state.Hostname, "127.0.0.1")
}

func Test_getIps(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/net/v1/nodes", r.URL.String())
		assert.Equal(t, "GET", r.Method)
		w.Write([]byte(`[{
      "updated": "2020-03-14T13:37:00.000Z",
      "public_ips": [
        "127.0.0.2"
      ],
      "private_ip": "127.0.0.1",
      "hostname": "ip-172-0-0-1"
    }]`))
	}))
	defer ts.Close()
	ctx := newContext(ts)

	data := <-getIps(ctx)
	wanted := map[string][]string{"127.0.0.1": {"127.0.0.2"}}
	assert.DeepEqual(t, data, wanted)
}

func newContext(ts *httptest.Server) *mock.Context {
	env := mock.NewEnvironment()
	ctx := mock.NewContext(env)
	cluster := config.NewCluster(nil)
	cluster.SetURL(ts.URL)
	cluster.Config().SetPath("testDir/")
	ctx.SetCluster(cluster)
	return ctx
}

const (
	testArg      = "--help"
	testEnvValue = "GO_TEST_PROCESS_ENV=true"
)

func Test_runCalicoCtl(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer ts.Close()
	ctx := newContext(ts)
	stdout, err := runCalicoCtl(fakeExecCommandSuccess, []string{testArg}, ctx, []string{testEnvValue})
	if err != nil {
		t.Error(err)
		return
	}

	// Check to make sure the stdout is returned properly
	stdoutStr := string(stdout)
	testOutput := fmt.Sprintf("%v | %v %v", testEnvValue, path.Join("testDir", "subcommands/dcos-core-cli/env/bin/calicoctl"), testArg)
	if stdoutStr != testOutput {
		t.Errorf("stdout mismatch:\n%s\n vs \n%s", stdoutStr, testOutput)
	}

}

// TestShellProcessSuccess is a method that is called as a substitute for a shell command,
// the GO_TEST_PROCESS flag ensures that if it is called as part of the test suite, it is
// skipped.
func TestShellProcessSuccess(t *testing.T) {
	if os.Getenv("GO_TEST_PROCESS") != "1" {
		return
	}
	// Print out the test value to stdout
	var envValue string
	env := os.Environ()
	for _, v := range env {
		if v == testEnvValue {
			envValue = v
		}
	}
	args := os.Args
	fmt.Fprintf(os.Stdout, "%v | %v", envValue, strings.Join(args[len(args)-2:], " "))
	os.Exit(0)
}

// fakeExecCommandSuccess is a function that initialises a new exec.Cmd, one which will
// simply call TestShellProcessSuccess rather than the command it is provided. It will
// also pass through the command and its arguments as an argument to TestShellProcessSuccess
func fakeExecCommandSuccess(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestShellProcessSuccess", "--", command}
	cs = append(cs, args...)
	arg := os.Args[0]
	cmd := exec.Command(arg, cs...)
	cmd.Env = []string{"GO_TEST_PROCESS=1"}
	return cmd
}
