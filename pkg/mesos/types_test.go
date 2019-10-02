package mesos

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIP(t *testing.T) {
	slave := &Slave{}
	slave.PID = "slave(1)@172.31.15.225:5051"

	slaveIP := slave.IP()

	assert.Equal(t, slaveIP, "172.31.15.225")
}
