package job

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
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

	require.Equal(t, "sleepy-test-docker", job.ID)
	require.Equal(t, "A job that sleeps", job.Description)

	// Run
	require.Equal(t, "echo 'Snoozing ...'; sleep 10; echo 'Awake now!'", job.Run.Cmd)
	require.Equal(t, float32(2), job.Run.Cpus)
	require.Equal(t, 32, job.Run.Mem)
	require.Equal(t, 10, job.Run.Disk)

	// Docker
	require.Equal(t, "alpine:latest", job.Run.Docker.Image)
	require.True(t, job.Run.Docker.ForcePullImage)
	require.True(t, job.Run.Docker.Privileged)
}
