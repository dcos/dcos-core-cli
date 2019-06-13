# attach

The `dcos task attach` attaches the terminal standard streams to a running container.

It is then able to stream the STDIN of the CLI process over to the container, as well as
streaming the STDOUT/STDERR of the container back to the CLI process.

The command only works for tasks that were launched with the Universal Container Runtime (UCR)
and with a TTY.

## Usage

``` terminal
dcos task attach [--no-stdin] <task-id>

  --no-stdin   Don't attach the stdin of the CLI to the task
```

The mandatory argument `<task-id>` indicates which task to attach to. It can be a full task ID,
a partial task ID, or a Unix shell wildcard pattern (eg. 'my-task*').

## Implementation

The command is implemented through the [Mesos Operator API v1](http://mesos.apache.org/documentation/latest/operator-http-api).

### Finding the task to attach to

To start with, the command needs to identify the task to attach to. In order to do that, it will make
a [GET_TASKS](http://mesos.apache.org/documentation/latest/operator-http-api/#get_tasks) call.

It then loops through the tasks and tries to find a match which is exactly `<task-id>`,
contains`<task-id>`, or matches the wildcard pattern of `<task-id>`.

If there is an exact match (the full ID was passed), or there were exactly 1 match, the command selects the given task. Otherwise, it fails with an error (`task <task-id> not found` or `multiple matches for <task-id>`).

### ATTACH_CONTAINER_INPUT

Unless `--no-stdin` is passed, the command attaches its STDIN through an [ATTACH_CONTAINER_INPUT](http://mesos.apache.org/documentation/latest/operator-http-api/#attach_container_input) call.

The first message sent is of type `CONTAINER_ID` and contains the container ID of the task found in the above step.

Afterwards, the command will continuously send 3 types of `PROCESS_IO` messages:

- `DATA` messages of type `STDIN`, which contain the data being read from STDIN.
- `CONTROL` messages of type `TTY_INFO` which contain the terminal size and are sent each time a SIGWINCH signal is received.
- `CONTROL` messages of type `HEARTBEAT` are being sent every 30 seconds in order to keep the connection alive.

The command will always send an initial pair of `TTY_INFO` messages to force a redraw of the remote
terminal. The first one being of size [0,0] and the second one being the actual size of the local terminal.

Data being read from STDIN is also buffered in order to detect a special `escape sequence` (`Ctrl+P`, `Ctrl+Q`).
When encountered, all connections are closed and the command exits with an exit code of `0`.

### ATTACH_CONTAINER_OUTPUT

The command attaches its STDOUT through an [ATTACH_CONTAINER_OUTPUT](http://mesos.apache.org/documentation/latest/operator-http-api/#attach_container_output) call.

It then expects 2 types of `PROCESS_IO` messages:

- `DATA` messages of type `STDOUT`.
- `DATA` messages of type `STDERR`.

The data is then continously written to the command STDOUT or STDERR.

### Waiting for container exit status

When the task was launched with the Mesos executor (it has a parent container), and unless the calls to
`ATTACH_CONTAINER_INPUT` or `ATTACH_CONTAINER_OUTPUT` were terminated because of a termination signal or
the escape sequence being detected, the command will make a `WAIT_CONTAINER` call in order to determine
an exit code.

The `WAIT_CONTAINER` returns an `exit_status`, which is the return value of `wait(2)`.

If the container was terminated because of a signal, the command exits with an exit code of `128 + signal_value`.

If the container terminated normally (that is, by calling exit(3), _exit(2), or by returning from main()),
the command exits with the same exit code.
