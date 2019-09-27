package mesos

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	mesos "github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/agent"
	agentcalls "github.com/mesos/mesos-go/api/v1/lib/agent/calls"
	"github.com/mesos/mesos-go/api/v1/lib/encoding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttach(t *testing.T) {
	containerID := mesos.ContainerID{
		Value: "my_container",
		Parent: &mesos.ContainerID{
			Value: "my_parent_container",
		},
	}

	messages := []agent.ProcessIO{
		processIOData(agent.ProcessIO_Data_STDOUT, []byte("Hello")),
		processIOData(agent.ProcessIO_Data_STDERR, []byte("[INFO] This is fine\n")),
		processIOData(agent.ProcessIO_Data_STDOUT, []byte(" world!")),
	}

	sender := agentcalls.SenderFunc(func(ctx context.Context, r agentcalls.Request) (mesos.Response, error) {
		var decoder encoding.Decoder

		switch r.Call().Type {

		case agent.Call_ATTACH_CONTAINER_OUTPUT:
			if assert.NotNil(t, r.Call().AttachContainerOutput) {
				assert.Equal(t, containerID.Value, r.Call().AttachContainerOutput.ContainerID.Value)
			}

			decoder = encoding.DecoderFunc(func(u encoding.Unmarshaler) error {
				msg, ok := u.(*agent.ProcessIO)
				require.True(t, ok)

				if len(messages) == 0 {
					return io.EOF
				}
				*msg = messages[0]
				messages = messages[1:]
				return nil
			})

		case agent.Call_WAIT_CONTAINER:
			if assert.NotNil(t, r.Call().WaitContainer) {
				assert.Equal(t, containerID.Value, r.Call().WaitContainer.ContainerID.Value)
			}

			decoder = encoding.DecoderFunc(func(u encoding.Unmarshaler) error {
				r, ok := u.(*agent.Response)
				require.True(t, ok)

				exitStatus := int32(0)
				r.WaitContainer = &agent.Response_WaitContainer{
					ExitStatus: &exitStatus,
				}
				return nil
			})

		default:
			return nil, fmt.Errorf("unexpected call type %s", r.Call().Type)
		}

		return &mesos.ResponseWrapper{
			Decoder: decoder,
		}, nil
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	opts := TaskIOOpts{
		Sender: sender,
		Stdout: &stdout,
		Stderr: &stderr,
	}

	taskIO, err := NewTaskIO(containerID, opts)
	require.NoError(t, err)

	exitCode, err := taskIO.Attach()
	require.NoError(t, err)

	assert.Equal(t, "Hello world!", stdout.String())
	assert.Equal(t, "[INFO] This is fine\n", stderr.String())
	assert.Equal(t, 0, exitCode)
}

func TestExec(t *testing.T) {
	containerID := mesos.ContainerID{
		Value: "my_container",
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	opts := TaskIOOpts{
		Stdout: &stdout,
		Stderr: &stderr,
	}

	taskIO, err := NewTaskIO(containerID, opts)
	require.NoError(t, err)

	messages := []agent.ProcessIO{
		processIOData(agent.ProcessIO_Data_STDOUT, []byte("Hello")),
		processIOData(agent.ProcessIO_Data_STDERR, []byte("[INFO] This is fine\n")),
		processIOData(agent.ProcessIO_Data_STDOUT, []byte(" world!")),
	}

	taskIO.opts.Sender = agentcalls.SenderFunc(func(ctx context.Context, r agentcalls.Request) (mesos.Response, error) {
		var decoder encoding.Decoder

		switch r.Call().Type {

		case agent.Call_LAUNCH_NESTED_CONTAINER_SESSION:

			call := r.Call().LaunchNestedContainerSession
			if assert.NotNil(t, call) {

				parentContainer := call.ContainerID.Parent
				if assert.NotNil(t, parentContainer) {
					assert.Equal(t, "my_container", parentContainer.Value)
				}

				cmdInfo := call.Command
				if assert.NotNil(t, cmdInfo) {
					assert.Equal(t, []string{"exit", "10"}, cmdInfo.Arguments)
				}
			}

			decoder = encoding.DecoderFunc(func(u encoding.Unmarshaler) error {
				msg, ok := u.(*agent.ProcessIO)
				require.True(t, ok)

				if len(messages) == 0 {
					return io.EOF
				}
				*msg = messages[0]
				messages = messages[1:]
				return nil
			})

		case agent.Call_WAIT_CONTAINER:
			if assert.NotNil(t, r.Call().WaitContainer) {
				assert.Equal(t, taskIO.containerID.Value, r.Call().WaitContainer.ContainerID.Value)
			}

			decoder = encoding.DecoderFunc(func(u encoding.Unmarshaler) error {
				r, ok := u.(*agent.Response)
				require.True(t, ok)

				exitStatus := int32(10 << 8) // exited with 10

				r.WaitContainer = &agent.Response_WaitContainer{
					ExitStatus: &exitStatus,
				}
				return nil
			})

		default:
			return nil, fmt.Errorf("unexpected call type %s", r.Call().Type)
		}

		return &mesos.ResponseWrapper{
			Decoder: decoder,
		}, nil
	})

	exitCode, err := taskIO.Exec("exit", "10")
	require.NoError(t, err)

	assert.Equal(t, "Hello world!", stdout.String())
	assert.Equal(t, "[INFO] This is fine\n", stderr.String())
	assert.Equal(t, 10, exitCode)
}

func processIOData(kind agent.ProcessIO_Data_Type, data []byte) agent.ProcessIO {
	return agent.ProcessIO{
		Type: agent.ProcessIO_DATA,
		Data: &agent.ProcessIO_Data{
			Type: kind,
			Data: data,
		},
	}
}
