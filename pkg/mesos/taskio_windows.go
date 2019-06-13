package mesos

import (
	mesos "github.com/mesos/mesos-go/api/v1/lib"
)

// handleSignals is a no-op on Windows.
func (t *TaskIO) handleSignals(stdinFd int, ttyInfoCh chan<- *mesos.TTYInfo, done chan<- struct{}) {}
