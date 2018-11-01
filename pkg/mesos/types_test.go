package mesos

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIP(t *testing.T) {
	slave := &Slave{}
	slave.PID = "slave(1)@172.31.15.225:5051"

	slaveIP := slave.IP()

	require.Equal(t, slaveIP, "172.31.15.225")
}
