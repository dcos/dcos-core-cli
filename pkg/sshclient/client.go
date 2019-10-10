package sshclient

import (
	"io"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

// Client to execute commands over SSH.
type Client struct {
	args   []string
	logger *logrus.Logger
	opts   ClientOpts
}

// ClientOpts defines all available options that can be set on a Client.
type ClientOpts struct {
	BinaryPath string
	Input      io.Reader
	Out        io.Writer
	ErrOut     io.Writer
	SSHOptions []string
	Config     string
	User       string
	Proxy      string
	Host       string
}

// -A	Enables forwarding of the authentication agent connection.
// -t	Force pseudo-terminal allocation. Used to execute arbitrary screen-based programs on a remote machine.
var baseArgs = []string{
	"-A",
	"-t",
}

// NewClient creates a new client to start a SSH session.
func NewClient(opts ClientOpts, logger *logrus.Logger) (*Client, error) {
	if opts.BinaryPath == "" {
		var err error
		opts.BinaryPath, err = exec.LookPath("ssh")
		if err != nil {
			return nil, err
		}
	}
	if opts.Input != nil {
		opts.Input = os.Stdin
	}
	if opts.Out != nil {
		opts.Out = os.Stdout
	}
	if opts.ErrOut != nil {
		opts.ErrOut = os.Stderr
	}

	c := &Client{opts: opts}
	c.logger = logger
	c.configureSSHOptions()
	c.configureDestination()

	return c, nil
}

func (c *Client) configureSSHOptions() {
	for _, option := range c.opts.SSHOptions {
		c.args = append(c.args, "-o", option)
	}

	if c.opts.Config != "" {
		c.args = append(c.args, "-F", c.opts.Config)
	}
}

func (c *Client) configureDestination() {
	c.logger.Debugf("Trying to establish connection to %s\n", c.opts.Host)
	if c.opts.Proxy != "" {
		c.logger.Debugf("Using %s as a proxy node\n", c.opts.Proxy)
		c.args = append(c.args, baseArgs...)

		if c.opts.Config == "" {
			c.args = append(c.args, "-l", c.opts.User)
		}

		c.args = append(c.args, c.opts.Proxy)
		c.args = append(c.args, "ssh")

		for _, option := range c.opts.SSHOptions {
			c.args = append(c.args, "-o", option)
		}
		if c.opts.Config == "" {
			c.args = append(c.args, "-l", c.opts.User)
		}
	}

	c.args = append(c.args, baseArgs...)
	c.args = append(c.args, c.opts.Host)
}

// Run adds the optional remote command and starts the SSH session.
func (c *Client) Run(command []string) error {
	// Escape remote commands to execute them on the target remote.
	if len(command) > 0 {
		command = append([]string{"'"}, command...)
		command = append(command, "'")
	}

	args := append(c.args, command...)
	cmd := exec.Command(c.opts.BinaryPath, args...) // nolint: gosec
	c.logger.Debugf("Running: %v\n", cmd.Args)

	cmd.Stdin = c.opts.Input
	cmd.Stdout = c.opts.Out
	cmd.Stderr = c.opts.ErrOut

	return cmd.Run()
}
