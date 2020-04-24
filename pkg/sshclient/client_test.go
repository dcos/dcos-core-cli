package sshclient

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
)

func TestNewClientArgs(t *testing.T) {
	logger := logrus.New()

	tests := []struct {
		name     string
		given    ClientOpts
		expected []string
	}{
		{name: "no opts", given: ClientOpts{}, expected: []string{"-A", "-t", ""}},
		{name: "given host", given: ClientOpts{Host: "192.0.2.1"}, expected: []string{"-A", "-t", "192.0.2.1"}},
		{name: "proxy", given: ClientOpts{Proxy: "192.0.2.1"}, expected: []string{"-A", "-t", "-l", "", "192.0.2.1", "ssh", "-l", "", "-A", "-t", ""}},
		{name: "config", given: ClientOpts{Config: "ssh.config"}, expected: []string{"-F", "ssh.config", "-A", "-t", ""}},
		{name: "user", given: ClientOpts{User: "root"}, expected: []string{"-A", "-t", ""}},
		{name: "shh options", given: ClientOpts{SSHOptions: []string{"-6", "-C", "-q"}}, expected: []string{"-o", "-6", "-o", "-C", "-o", "-q", "-A", "-t", ""}},
		{name: "user and proxy", given: ClientOpts{User: "root", Proxy: "192.0.2.1"}, expected: []string{"-A", "-t", "-l", "root", "192.0.2.1", "ssh", "-l", "root", "-A", "-t", ""}},
		{name: "config and proxy", given: ClientOpts{Config: "ssh.config", Proxy: "192.0.2.1"}, expected: []string{"-F", "ssh.config", "-A", "-t", "192.0.2.1", "ssh", "-A", "-t", ""}},
		{name: "user and config and proxy", given: ClientOpts{User: "root", Config: "ssh.config", Proxy: "192.0.2.1"}, expected: []string{"-F", "ssh.config", "-A", "-t", "192.0.2.1", "ssh", "-A", "-t", ""}},
		{name: "options and proxy", given: ClientOpts{SSHOptions: []string{"-6", "-C", "-q"}, Proxy: "192.0.2.1"}, expected: []string{"-o", "-6", "-o", "-C", "-o", "-q", "-A", "-t", "-l", "", "192.0.2.1", "ssh", "-o", "-6", "-o", "-C", "-o", "-q", "-l", "", "-A", "-t", ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClient(tt.given, logger)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got.args)
		})
	}
}

func TestNewClientWithNoSshOnPath(t *testing.T) {
	path, exist := os.LookupEnv("PATH")
	require.True(t, exist)
	err := os.Setenv("PATH", "")
	require.NoError(t, err)
	defer func() { _ = os.Setenv("PATH", path) }()

	client, err := NewClient(ClientOpts{}, nil)
	assert.Nil(t, client)
	assert.Error(t, err)
}

func TestNewClientSetDefault(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetLevel(logrus.DebugLevel)
	client, err := NewClient(ClientOpts{}, logger)
	require.NoError(t, err)
	assert.Equal(t, &Client{
		logger: logger,
		args:   []string{"-A", "-t", ""},
		opts: ClientOpts{
			BinaryPath: "/usr/bin/ssh",
		},
	}, client)
	assert.Contains(t,  buf.String(), "Trying to establish connection to")
}

func TestNewClientOverrideInputOutAndErrorOutWhenSetToStd(t *testing.T) {
	logger := logrus.New()
	var buf *bytes.Buffer
	client, err := NewClient(ClientOpts{
		Input:  strings.NewReader(""),
		Out:    buf,
		ErrOut: buf,
	}, logger)
	require.NoError(t, err)
	assert.Equal(t, &Client{
		logger: logger,
		args:   []string{"-A", "-t", ""},
		opts: ClientOpts{
			BinaryPath: "/usr/bin/ssh",
			Input:      os.Stdin,
			Out:        os.Stdout,
			ErrOut:     os.Stderr,
		},
	}, client)
}

func TestPrepareCommandEmptyCommand(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetLevel(logrus.DebugLevel)
	opts := ClientOpts{
		Config: "ssh.config",
		Host: "192.0.2.2",
		Proxy: "192.0.2.1",
		SSHOptions: []string{"-6", "-C", "-q"},
		User: "root",
	}
	c, err := NewClient(opts, logger)
	require.NoError(t, err)

	cmd := c.prepareCommand(nil)
	assert.Equal(t, []string{"/usr/bin/ssh", "-o", "-6", "-o", "-C", "-o", "-q", "-F", "ssh.config", "-A", "-t", "192.0.2.1", "ssh", "-o", "-6", "-o", "-C", "-o", "-q", "-A", "-t", "192.0.2.2"}, cmd.Args)
	cmd = c.prepareCommand([]string{"/bin/bash", "'-c'", "echo \"OK\""})
	assert.Equal(t, []string{"/usr/bin/ssh", "-o", "-6", "-o", "-C", "-o", "-q", "-F", "ssh.config", "-A", "-t", "192.0.2.1", "ssh", "-o", "-6", "-o", "-C", "-o", "-q", "-A", "-t", "192.0.2.2", "'", "/bin/bash", "'-c'", "echo \"OK\"", "'"}, cmd.Args)
	assert.Equal(t, opts.Input, cmd.Stdin)
	assert.Equal(t, opts.Out, cmd.Stdout)
	assert.Equal(t, opts.ErrOut, cmd.Stderr)
	assert.Contains(t, buf.String(), "Running: [/usr/bin/ssh -o -6 -o -C -o -q -F ssh.config -A -t 192.0.2.1 ssh -o -6 -o -C -o -q -A -t 192.0.2.2]")
}
