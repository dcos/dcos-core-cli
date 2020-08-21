package calico

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dcos/dcos-cli/pkg/config"
	"github.com/dcos/dcos-cli/pkg/mock"
)

func Test_runCalicoCtl(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	calicoCtlPath := path.Join(os.TempDir(), "subcommands/dcos-core-cli/env/bin/calicoctl")
	tests := []struct {
		args     []string
		level    logrus.Level
		expected *exec.Cmd
	}{
		{nil, logrus.DebugLevel, exec.Command(calicoCtlPath, "--help")},
		{[]string{"--help"}, logrus.DebugLevel, exec.Command(calicoCtlPath, "-l", "debug", "--help")},
		{[]string{"version"}, logrus.InfoLevel, exec.Command(calicoCtlPath, "-l", "info", "version")},
		{[]string{"version", "-h"}, logrus.PanicLevel, exec.Command(calicoCtlPath, "-l", "panic", "version", "-h")},
	}
	for _, tt := range tests {
		ctx, _ := newContext(ts)
		ctx.Logger().SetLevel(tt.level)
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			cmd := runCalicoCtl(tt.args, ctx, []string{"A=1", "B=2"})
			assert.Equal(t, tt.expected.Path, cmd.Path)
			assert.Equal(t, tt.expected.Args, cmd.Args)
			assert.Equal(t, append(os.Environ(), "A=1", "B=2"), cmd.Env)
			assert.Equal(t, ctx.ErrOut(), cmd.Stderr)
			assert.Equal(t, ctx.Input(), cmd.Stdin)
			assert.Equal(t, ctx.Out(), cmd.Stdout)
		})
	}
}

func TestGetEnvForEnterprise(t *testing.T) {
	grpcServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get("Authorization"))
	}))
	cacert, err := ioutil.TempFile("", "*dcos-ca.crt")
	require.NoError(t, err)
	defer os.Remove(cacert.Name())
	// https://github.com/golang/go/blob/go1.15/src/net/http/internal/testcert.go#L13
	_, err = cacert.Write([]byte(`-----BEGIN CERTIFICATE-----
MIICEzCCAXygAwIBAgIQMIMChMLGrR+QvmQvpwAU6zANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMCAXDTcwMDEwMTAwMDAwMFoYDzIwODQwMTI5MTYw
MDAwWjASMRAwDgYDVQQKEwdBY21lIENvMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCB
iQKBgQDuLnQAI3mDgey3VBzWnB2L39JUU4txjeVE6myuDqkM/uGlfjb9SjY1bIw4
iA5sBBZzHi3z0h1YV8QPuxEbi4nW91IJm2gsvvZhIrCHS3l6afab4pZBl2+XsDul
rKBxKKtD1rGxlG4LjncdabFn9gvLZad2bSysqz/qTAUStTvqJQIDAQABo2gwZjAO
BgNVHQ8BAf8EBAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUw
AwEB/zAuBgNVHREEJzAlggtleGFtcGxlLmNvbYcEfwAAAYcQAAAAAAAAAAAAAAAA
AAAAATANBgkqhkiG9w0BAQsFAAOBgQCEcetwO59EWk7WiJsG4x8SY+UIAA+flUI9
tyC4lNhbcF2Idq9greZwbYCqTTTr2XiRNSMLCOjKyI7ukPoPjo16ocHj+P3vZGfs
h1fIw3cSS2OolhloGw/XM6RWPWtPAlGykKLciQrBru5NAPvCMsb/I1DAceTiotQM
fblo6RBxUQ==
-----END CERTIFICATE-----`))
	require.NoError(t, err)

	err = os.Setenv("DCOS_ACS_TOKEN", "12345")
	require.NoError(t, err)
	defer os.Unsetenv("DCOS_ACS_TOKEN")
	err = os.Setenv("DCOS_TLS_CA_PATH", cacert.Name())
	require.NoError(t, err)
	defer os.Unsetenv("DCOS_TLS_CA_PATH")

	u, err := url.Parse(grpcServer.URL)
	require.NoError(t, err)
	_, grpcPort, err := net.SplitHostPort(u.Host)
	require.NoError(t, err)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/mesos_dns/v1/hosts/leader.mesos":
			w.Write([]byte(`[{"host": "leader.mesos.", "ip": "192.0.2.3"}]`))
			return
		case "/net/v1/nodes":
			w.Write([]byte(`
				[
				   {"public_ips":[ "18.207.110.46"],"private_ip":"192.0.2.1"},
				   {"public_ips":[], "private_ip":"192.0.2.2"},
				   {"public_ips":["127.0.0.1"],"private_ip":"192.0.2.3"}
				]`))
			return
		case "/dcos-metadata/dcos-version.json":
			w.Write([]byte(`{"dcos-variant": "enterprise"}`))
			return
		case "/ca/dcos-ca.crt":
			w.Write([]byte(`CERTIFICATE`))
			return
		}
		t.Error("path is not supported: " + r.URL.Path)
	}))
	defer ts.Close()
	defer grpcServer.Close()
	ctx, out := newContext(ts)

	env, err := getEnvironment(ctx, ":"+grpcPort)
	assert.NoError(t, err)
	assert.Equal(t, []string{
		"ETCD_CUSTOM_GRPC_METADATA=authorization:token=",
		"ETCD_ENDPOINTS=127.0.0.1:" + grpcPort,
		"ETCD_CA_CERT_FILE=" + cacert.Name(),
	}, env)

	output, err := ioutil.ReadAll(out)
	assert.NoError(t, err)
	assert.Equal(t, `level=debug msg="Get leader private IP"
level=debug msg="Get nodes public IPs"
`, string(output))

}

func TestGetEnvForOSS(t *testing.T) {
	grpcServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	u, err := url.Parse(grpcServer.URL)
	require.NoError(t, err)
	_, grpcPort, err := net.SplitHostPort(u.Host)
	require.NoError(t, err)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/mesos_dns/v1/hosts/leader.mesos":
			w.Write([]byte(`[{"host": "leader.mesos.", "ip": "192.0.2.3"}]`))
			return
		case "/net/v1/nodes":
			w.Write([]byte(`
				[
				   {"public_ips":[ "18.207.110.46"],"private_ip":"192.0.2.1"},
				   {"public_ips":[], "private_ip":"192.0.2.2"},
				   {"public_ips":["127.0.0.1"],"private_ip":"192.0.2.3"}
				]`))
			return
		case "/dcos-metadata/dcos-version.json":
			w.Write([]byte(`{"dcos-variant": "oss"}`))
			return
		}
		t.Error("path is not supported: " + r.URL.Path)
	}))
	defer ts.Close()
	defer grpcServer.Close()
	ctx, out := newContext(ts)

	env, err := getEnvironment(ctx, ":"+grpcPort)
	assert.NoError(t, err)
	assert.Equal(t, []string{"ETCD_CUSTOM_GRPC_METADATA=authorization:token=", "ETCD_ENDPOINTS=127.0.0.1:" + grpcPort}, env)

	output, err := ioutil.ReadAll(out)
	assert.NoError(t, err)
	assert.Equal(t, `level=debug msg="Get leader private IP"
level=debug msg="Get nodes public IPs"
`, string(output))

}

func newContext(ts *httptest.Server) (*mock.Context, *bytes.Buffer) {
	out := new(bytes.Buffer)
	env := mock.NewEnvironment()
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
