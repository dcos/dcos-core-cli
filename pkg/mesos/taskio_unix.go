// +build linux darwin

package mesos

import (
	"os"
	"os/signal"
	"syscall"

	mesos "github.com/mesos/mesos-go/api/v1/lib"
)

// handleSignals registers a handler for the SIGWINCH, SIGINT and SIGTERM signals.
//
// For each SIGWINCH signal it will send the new terminal size in the ttyInfoCh channel.
//
// The ttyInfoCh channel is connected to the request body of the `ATTACH_CONTAINER_INPUT` call
// and sends the appropriate TTY info control messages to Mesos.
//
// It also makes sure the input connection is closed properly when receiving SIGINT/SIGTERM signals.
// See https://issues.apache.org/jira/browse/MESOS-9838
func (t *TaskIO) handleSignals(stdinFd int, ttyInfoCh chan<- *mesos.TTYInfo, done chan<- struct{}) {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGWINCH, syscall.SIGINT, syscall.SIGTERM)

	for sig := range sigs {
		switch sig {
		case syscall.SIGWINCH:
			ttyInfo, err := t.ttyInfo(stdinFd)
			if err != nil {
				// TODO(bamarni): log?
				continue
			}
			ttyInfoCh <- ttyInfo

		case syscall.SIGINT, syscall.SIGTERM:
			t.terminationSignalDetected = true
			close(done)
		}
	}
}
