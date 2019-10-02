package job

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
)

func TestParseJSONJob(t *testing.T) {
	reader := strings.NewReader(`
{
	"id": "sleepy-test-docker",
	"description": "A job that sleeps",
	"run": {
		"cmd": "echo 'Snoozing ...'; sleep 10; echo 'Awake now!'",
		"cpus": 2,
		"mem": 32,
		"disk": 10,
		"docker": {
			"image": "alpine:latest",
			"forcePullImage": true,
			"privileged": true
		}
	}
}`)

	job, err := parseJSONJob(reader)
	require.NoError(t, err)

	assert.Equal(t, "sleepy-test-docker", job.ID)
	assert.Equal(t, "A job that sleeps", job.Description)

	// Run
	assert.Equal(t, "echo 'Snoozing ...'; sleep 10; echo 'Awake now!'", job.Run.Cmd)
	assert.Equal(t, float32(2), job.Run.Cpus)
	assert.Equal(t, 32, job.Run.Mem)
	assert.Equal(t, 10, job.Run.Disk)

	// Docker
	assert.Equal(t, "alpine:latest", job.Run.Docker.Image)
	assert.True(t, job.Run.Docker.ForcePullImage)
	assert.True(t, job.Run.Docker.Privileged)
}
