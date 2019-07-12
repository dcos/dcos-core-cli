package mesos

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/docker/docker/pkg/term"
	mesos "github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/agent"
	agentcalls "github.com/mesos/mesos-go/api/v1/lib/agent/calls"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

var defaultHeartbeatInterval = 30 * time.Second
var defaultEscapeSequence = []byte{0x10, 0x11} // CTRL-P, CTRL-Q
var defaultTermValue = "xterm"                 // default value for the TERM env var

// TaskIOOpts are options for a TaskIO.
type TaskIOOpts struct {
	Stdin             io.Reader
	Stdout            io.Writer
	Stderr            io.Writer
	Interactive       bool
	TTY               bool
	User              string
	HeartbeatInterval time.Duration
	EscapeSequence    []byte
	Sender            agentcalls.Sender
	Logger            *logrus.Logger
}

// TaskIO is an abstraction used to stream I/O between a running Mesos task and the local terminal.
//
// A TaskIO object can only be used for a single streaming session (through Attach or Exec),
// for subsequent streaming sessions one should instanciate new TaskIO objects.
type TaskIO struct {
	containerID mesos.ContainerID
	opts        TaskIOOpts

	// This channel is used to signal when we're attached to the container output.
	outputAttached chan struct{}

	// This cancellable context will be shared across the different HTTP calls,
	// when any HTTP request finishes the context will be cancelled. We do not
	// have a rety logic.
	ctx        context.Context
	cancelFunc context.CancelFunc

	// Channel to receive errors from the different goroutines.
	errCh chan error

	// We use a WaitGroup and will wait for all HTTP connections to be closed before returning.
	wg sync.WaitGroup

	exitSequenceDetected      bool
	terminationSignalDetected bool
}

// NewTaskIO creates a new TaskIO.
func NewTaskIO(containerID mesos.ContainerID, opts TaskIOOpts) (*TaskIO, error) {
	if opts.Stdin == nil {
		opts.Stdin = os.Stdin
	}
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}
	if opts.HeartbeatInterval == 0 {
		opts.HeartbeatInterval = defaultHeartbeatInterval
	}
	if len(opts.EscapeSequence) == 0 {
		opts.EscapeSequence = defaultEscapeSequence
	}

	ctx, cancel := context.WithCancel(context.TODO())

	return &TaskIO{
		containerID:    containerID,
		outputAttached: make(chan struct{}),
		opts:           opts,
		ctx:            ctx,
		cancelFunc:     cancel,
		errCh:          make(chan error, 2),
	}, nil
}

// Exec launches a nested task based on the given command and attaches the stdin/stdout/stderr of the CLI
// to the STDIN/STDOUT/STDERR of the nested container.
func (t *TaskIO) Exec(cmd string, args ...string) (int, error) {
	defer t.cancelFunc()
	defer close(t.errCh)

	// Launch nested container session.
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		defer t.cancelFunc()

		err := t.launchNestedContainerSession(cmd, args...)
		if err != nil && t.ctx.Err() == nil {
			t.errCh <- err
		}
	}()

	if t.opts.Interactive {
		t.attachInput()
	}
	return t.wait()
}

// Attach attaches the stdin/stdout/stderr of the CLI to the STDIN/STDOUT/STDERR of a running task.
//
// As of now, we can only attach to tasks launched with a remote TTY already set up for them.
// If we try to attach to a task that was launched without a remote TTY attached, an error is returned.
func (t *TaskIO) Attach() (int, error) {
	defer t.cancelFunc()
	defer close(t.errCh)

	// Attach container outputs to the CLI stdout/stderr.
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		defer t.cancelFunc()

		err := t.attachContainerOutput()
		if err != nil && t.ctx.Err() == nil {
			t.errCh <- err
		}
	}()

	if t.opts.Interactive {
		t.attachInput()
	}
	return t.wait()
}

// launchNestedContainerSession launches a nested container session.
func (t *TaskIO) launchNestedContainerSession(cmd string, args ...string) error {
	call := t.launchNestedContainerSessionCall(cmd, args...)

	// Send returns immediately with a Response from which output may be decoded.
	resp, err := t.opts.Sender.Send(t.ctx, agentcalls.NonStreaming(call))
	if resp != nil {
		defer resp.Close()
	}
	if err != nil {
		close(t.outputAttached)
		return err
	}
	close(t.outputAttached)

	return t.forwardContainerOutput(resp)
}

// launchNestedContainerSessionCall returns the request payload for the LAUNCH_NESTED_CONTAINER_SESSION call.
func (t *TaskIO) launchNestedContainerSessionCall(cmd string, args ...string) *agent.Call {

	// Override the container ID with the current container ID as the parent, and generate
	// a new UUID for the nested container used to run commands passed to `task exec`.
	parentContainerID := t.containerID
	t.containerID = mesos.ContainerID{
		Parent: &parentContainerID,
		Value:  uuid.NewV4().String(),
	}

	t.opts.Logger.Infof("Launching nested container with ID '%s'", t.containerID.Value)

	shell := false

	cmdInfo := &mesos.CommandInfo{
		Value:     &cmd,
		Arguments: append([]string{cmd}, args...),
		Shell:     &shell,
	}

	if t.opts.User != "" {
		cmdInfo.User = &t.opts.User
	}

	containerInfo := &mesos.ContainerInfo{
		Type: mesos.ContainerInfo_MESOS.Enum(),
	}

	if t.opts.TTY {
		containerInfo.TTYInfo = &mesos.TTYInfo{}

		cmdInfo.Environment = &mesos.Environment{
			Variables: []mesos.Environment_Variable{
				{
					Name:  "TERM",
					Type:  mesos.Environment_Variable_VALUE.Enum(),
					Value: &defaultTermValue,
				},
			},
		}
	}
	return agentcalls.LaunchNestedContainerSession(t.containerID, cmdInfo, containerInfo)
}

// attachInput handles the attachContainerInput call in a goroutine once the container output is attached.
func (t *TaskIO) attachInput() {
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		// Wait for the output to be attached before attaching the input.
		<-t.outputAttached

		err := t.attachContainerInput()
		if err != nil && t.ctx.Err() == nil {
			t.cancelFunc()
			t.errCh <- err
		}
	}()
}

// attachContainerInput streams the STDIN of the CLI to the remote container.
// It also sends TTYInfo messages whenever the terminal is resized and
// heartbeats messages every 30 seconds.
func (t *TaskIO) attachContainerInput() error {

	// Channels for window resize and termination signals.
	ttyInfoCh := make(chan *mesos.TTYInfo, 2)
	receivedTerminationSignal := make(chan struct{})

	if t.opts.TTY {
		stdinFd, err := t.stdinFd()
		if err != nil {
			return err
		}

		if !terminal.IsTerminal(stdinFd) {
			return fmt.Errorf("stdin is not a terminal")
		}

		// Create a proxy reader for stdin which is able to detect the escape sequence.
		t.opts.Stdin = term.NewEscapeProxy(t.opts.Stdin, t.opts.EscapeSequence)

		// Set the terminal in raw mode and make sure it's restored
		// to its previous state before the function returns.
		oldState, err := terminal.MakeRaw(stdinFd)
		if err != nil {
			return err
		}
		defer terminal.Restore(stdinFd, oldState)

		if runtime.GOOS != "windows" {
			// To force a redraw of the remote terminal, we first resize it to 0 before setting it
			// to the actual size of our local terminal. After that, all terminal resizing is handled
			// in our SIGWINCH handler.
			ttyInfoCh <- &mesos.TTYInfo{WindowSize: &mesos.TTYInfo_WindowSize{}}
			ttyInfo, err := t.ttyInfo(stdinFd)
			if err != nil {
				return err
			}
			ttyInfoCh <- ttyInfo
		}

		go t.handleSignals(stdinFd, ttyInfoCh, receivedTerminationSignal)
	}

	// Must be buffered to avoid blocking below.
	aciCh := make(chan *agent.Call, 1)

	// Very first input message MUST be this.
	aciCh <- agentcalls.AttachContainerInput(t.containerID)

	go func() {
		defer close(aciCh)

		input := make(chan []byte)
		go func() {
			defer func() {
				// Push an empty string to indicate EOF to the server and close
				// the input channel to signal that we are done processing input.
				input <- []byte("")
				close(input)
			}()

			for {
				bufLen := 512
				if t.opts.TTY {
					// TODO(bamarni): investigate on why "test_task:test_attach" fails if the buffer size below
					// is greater than 1, this looks like an issue with the term package we're using, which doesn't
					// work properly when reading more than 1 char at a time.
					bufLen = 1
				}
				buf := make([]byte, bufLen)
				n, err := t.opts.Stdin.Read(buf)
				if _, ok := err.(term.EscapeError); ok {
					t.cancelFunc()
					t.exitSequenceDetected = true
					return
				}
				if n > 0 {
					select {
					case input <- buf[:n]:
					case <-t.ctx.Done():
						return
					}
				}
				// TODO(jdef) check for temporary error?
				if err != nil {
					return
				}
			}
		}()

		// create a ticker which will be used to send hearbeats
		// at given intervals in order to keep the connection alive.
		ticker := time.NewTicker(t.opts.HeartbeatInterval)
		defer ticker.Stop()

		for {
			select {
			case <-t.ctx.Done():
				return

			case <-receivedTerminationSignal:
				return

			case ttyInfo, ok := <-ttyInfoCh:
				if !ok {
					return
				}
				c := agentcalls.AttachContainerInputTTY(ttyInfo)

				select {
				case aciCh <- c:
				case <-t.ctx.Done():
					return
				}

			case data, ok := <-input:
				if !ok {
					return
				}
				c := agentcalls.AttachContainerInputData(data)

				select {
				case aciCh <- c:
				case <-t.ctx.Done():
					return
				}

			case <-ticker.C:
				c := agentcalls.AttachContainerInputHeartbeat(&agent.ProcessIO_Control_Heartbeat{
					Interval: &mesos.DurationInfo{
						Nanoseconds: t.opts.HeartbeatInterval.Nanoseconds(),
					},
				})
				select {
				case aciCh <- c:
				case <-t.ctx.Done():
					return
				}
			}
		}
	}()

	acif := agentcalls.FromChan(aciCh)

	err := agentcalls.SendNoData(t.ctx, t.opts.Sender, acif)
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

// attachContainerOutput attaches the CLI output to container stdout/stderr.
func (t *TaskIO) attachContainerOutput() error {

	// Send returns immediately with a Response from which output may be decoded.
	call := agentcalls.NonStreaming(agentcalls.AttachContainerOutput(t.containerID))
	resp, err := t.opts.Sender.Send(t.ctx, call)
	if resp != nil {
		defer resp.Close()
	}
	if err != nil {
		close(t.outputAttached)
		return err
	}
	close(t.outputAttached)

	return t.forwardContainerOutput(resp)
}

// wait waits for the streaming session to terminate and returns the appropriate exit code.
func (t *TaskIO) wait() (int, error) {
	t.wg.Wait()

	select {
	case err := <-t.errCh:
		return 0, err
	default:
		if t.exitSequenceDetected {
			return 0, nil
		}
		if t.terminationSignalDetected {
			// TODO(bamarni): maybe exit with the signal value + 128
			return 1, nil
		}
		return t.waitContainer()
	}
}

// waitContainer waits for the container to terminate and returns its exit code.
//
// The WAIT_CONTAINER call is a long-running call, however it doesn't work well on DC/OS
// as Admin Router returns with a 504 Gateway timeout after 60 seconds of inactivity.
//
// Theoretically the call should be done concurrently with ATTACH_CONTAINER_INPUT/ATTACH_CONTAINER_OUTPUT,
// but because of the aforementioned limitation, we simply make it once the ATTACH_CONTAINER_* calls terminate.
// While it works in practice, this creates a timing condition, the Mesos containerizer does not keep status
// of terminated containers, calling `wait_container` after container termination returns `Not found`.
//
// See https://jira.mesosphere.com/browse/DCOS_OSS-5282
func (t *TaskIO) waitContainer() (int, error) {

	// We are only able to get the 'exit_status' of tasks launched via the default executor
	// (i.e. as pods rather than via the command executor). In the future, mesos will deprecate
	// the command executor in favor of the default executor, so this check will  be able to go away.
	// In the meantime, we will always return '0' for tasks launched via the command executor.
	if t.containerID.GetParent() == nil {
		return 0, nil
	}

	// Once the ATTACH_CONTAINER_INPUT/ATTACH_CONTAINER_OUTPUT calls are terminated,
	// wait up to 30 seconds for the container exit status.
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	call := agentcalls.NonStreaming(agentcalls.WaitContainer(t.containerID))
	resp, err := t.opts.Sender.Send(ctx, call)
	if resp != nil {
		defer resp.Close()
	}
	if err != nil {
		return 0, err
	}

	var agentResp agent.Response
	err = resp.Decode(&agentResp)
	if err != nil {
		return 0, err
	}

	// Note:`exit_status` is the return value of `wait(2)`.
	// Callers must use the `wait(2)` family of macros to extract whether
	// the process exited cleanly and what the exit code was.
	//
	// The code below first checks for a terminating signal. If any, return
	// the signal value + 128. Otherwise return with the exit code.
	//
	// See https://github.com/bminor/glibc/blob/master/bits/waitstatus.h
	exitStatus := int(agentResp.GetWaitContainer().GetExitStatus())
	termSignal := exitStatus & 0x7f
	if termSignal > 0 {
		return termSignal + 128, nil
	}
	return exitStatus >> 8, nil
}

// stdinFd returns the file descriptor of stdin.
func (t *TaskIO) stdinFd() (int, error) {
	stdin, ok := t.opts.Stdin.(*os.File)
	if !ok {
		return 0, errors.New("stdin is not a file")
	}

	return int(stdin.Fd()), nil
}

// ttyInfo returns a Mesos TTYInfo struct with the current terminal size.
func (t *TaskIO) ttyInfo(fd int) (*mesos.TTYInfo, error) {
	w, h, err := terminal.GetSize(fd)
	if err != nil {
		return nil, err
	}
	return &mesos.TTYInfo{WindowSize: &mesos.TTYInfo_WindowSize{
		Rows:    uint32(h),
		Columns: uint32(w),
	}}, nil
}

// forwardContainerOutput forwards output of the LAUNCH_NESTED_CONTAINER_SESSION
// or ATTACH_CONTAINER_OUTPUT responses to STDOUT/STDERR.
func (t *TaskIO) forwardContainerOutput(resp mesos.Response) error {
	forward := func(b []byte, out io.Writer) error {
		n, err := out.Write(b)
		if err == nil && len(b) != n {
			err = io.ErrShortWrite
		}
		return err
	}

	for {
		var pio agent.ProcessIO
		err := resp.Decode(&pio)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		switch pio.GetType() {
		case agent.ProcessIO_DATA:
			data := pio.GetData()
			switch data.GetType() {
			case agent.ProcessIO_Data_STDOUT:
				if err := forward(data.GetData(), t.opts.Stdout); err != nil {
					return err
				}
			case agent.ProcessIO_Data_STDERR:
				if err := forward(data.GetData(), t.opts.Stderr); err != nil {
					return err
				}
			}
		}
	}
}
