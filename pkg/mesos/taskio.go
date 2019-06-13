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
	"golang.org/x/crypto/ssh/terminal"
)

var defaultHeartbeatInterval = 30 * time.Second
var defaultEscapeSequence = []byte{0x10, 0x11} // CTRL-P, CTRL-Q

// TaskIOOpts are options for a TaskIO.
type TaskIOOpts struct {
	Stdin             io.Reader
	Stdout            io.Writer
	Stderr            io.Writer
	Interactive       bool
	TTY               bool
	HeartbeatInterval time.Duration
	EscapeSequence    []byte
	Sender            agentcalls.Sender
}

// TaskIO is an abstraction used to stream I/O between a running Mesos task and the local terminal.
type TaskIO struct {
	containerID               mesos.ContainerID
	opts                      TaskIOOpts
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
	return &TaskIO{
		containerID: containerID,
		opts:        opts,
	}, nil
}

// Attach attaches the stdin/stdout/stderr of the CLI to the STDIN/STDOUT/STDERR of a running task.
//
// As of now, we can only attach to tasks launched with a remote TTY already set up for them.
// If we try to attach to a task that was launched without a remote TTY attached, an error is returned.
func (t *TaskIO) Attach() (int, error) {

	// This cancellable context will be shared across the different HTTP calls,
	// when any HTTP request finishes the context will be cancelled. We do not
	// have a rety logic.
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	// Channel to receive errors from the ATTACH_CONTAINER_INPUT/ATTACH_CONTAINER_OUTPUT goroutines.
	errCh := make(chan error, 2)
	defer close(errCh)

	// We use a WaitGroup and will wait for all HTTP connections to be closed before returning.
	var wg sync.WaitGroup

	// When the input is interactive, attach to the container input.
	if t.opts.Interactive {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer cancel()

			err := t.attachContainerInput(ctx)
			if err != nil && ctx.Err() == nil {
				errCh <- err
			}
		}()
	}

	// Attach container outputs to the CLI stdout/stderr.
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()

		err := t.attachContainerOutput(ctx)
		if err != nil && ctx.Err() == nil {
			errCh <- err
		}
	}()

	wg.Wait()

	select {
	case err := <-errCh:
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

// attachContainerInput streams the STDIN of the CLI to the remote container.
// It also sends TTYInfo messages whenever the terminal is resized and
// heartbeats messages every 30 seconds.
func (t *TaskIO) attachContainerInput(ctx context.Context) error {
	stdinFd, err := t.stdinFd()
	if err != nil {
		return err
	}

	if !terminal.IsTerminal(stdinFd) {
		return fmt.Errorf("stdin is not a terminal")
	}

	// Set the terminal in raw mode and make sure it's restored
	// to its previous state before the function returns.
	oldState, err := terminal.MakeRaw(stdinFd)
	if err != nil {
		return err
	}
	defer terminal.Restore(stdinFd, oldState)

	// Create a proxy reader for stdin which is able to detect the escape sequence.
	t.opts.Stdin = term.NewEscapeProxy(t.opts.Stdin, t.opts.EscapeSequence)

	// Channels for window resize and termination signals.
	ttyInfoCh := make(chan *mesos.TTYInfo, 2)
	receivedTerminationSignal := make(chan struct{})

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

	// Must be buffered to avoid blocking below.
	aciCh := make(chan *agent.Call, 1)

	// Very first input message MUST be this.
	aciCh <- agentcalls.AttachContainerInput(t.containerID)

	go func() {
		defer close(aciCh)

		input := make(chan []byte)
		go func() {
			defer close(input)

			for {
				buf := make([]byte, 512) // not efficient to always do this
				n, err := t.opts.Stdin.Read(buf)
				if _, ok := err.(term.EscapeError); ok {
					t.exitSequenceDetected = true
					return
				}
				if n > 0 {
					select {
					case input <- buf[:n]:
					case <-ctx.Done():
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
			case <-ctx.Done():
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
				case <-ctx.Done():
					return
				}

			case data, ok := <-input:
				if !ok {
					return
				}
				c := agentcalls.AttachContainerInputData(data)

				select {
				case aciCh <- c:
				case <-ctx.Done():
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
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	acif := agentcalls.FromChan(aciCh)

	err = agentcalls.SendNoData(ctx, t.opts.Sender, acif)
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

// attachContainerOutput attaches the CLI output to container stdout/stderr.
func (t *TaskIO) attachContainerOutput(ctx context.Context) error {

	// Send returns immediately with a Response from which output may be decoded.
	call := agentcalls.NonStreaming(agentcalls.AttachContainerOutput(t.containerID))
	resp, err := t.opts.Sender.Send(ctx, call)
	if resp != nil {
		defer resp.Close()
	}
	if err != nil {
		return err
	}

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
