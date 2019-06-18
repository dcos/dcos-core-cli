# dcos task exec

The `dcos task exec` command runs a command inside a task.

It works by creating a nested container into the task, then attaching the standard streams as the
`dcos task attach` command would do.

## Usage

``` terminal
dcos task exec [-i|--interactive] [-t|--tty] [-u|--user=] <task-id> [--] <cmd> [<args>...]

  -i, --interactive   Attach a STDIN stream to the remote command for an interactive session
  -t, --tty           Attach a tty to the remote stream.
  -u, --user string   Run as the given user
```

The mandatory argument `<task-id>` indicates which task to attach to. It can be a full task ID,
a partial task ID, or a Unix shell wildcard pattern (eg. 'my-task*').

The CLI launches a nested container with the provided `<cmd>` and, optionally, `<args>`.

## Implementation

The command is implemented through the [Mesos Operator API v1](http://mesos.apache.org/documentation/latest/operator-http-api).

### Finding the task to attach to

To start with, the command needs to identify the task to attach to. In order to do that, it will make
a [GET_TASKS](http://mesos.apache.org/documentation/latest/operator-http-api/#get_tasks) call.

It then loops through the tasks and tries to find a match which is exactly `<task-id>`,
contains`<task-id>`, or matches the wildcard pattern of `<task-id>`.

If there is an exact match (the full ID was passed), or there were exactly 1 match, the command selects the given task. Otherwise, it fails with an error (`task <task-id> not found` or `multiple matches for <task-id>`).

### LAUNCH_NESTED_CONTAINER_SESSION

The command first launches a nested `MESOS` container using the [LAUNCH_NESTED_CONTAINER_SESSION](http://mesos.apache.org/documentation/latest/operator-http-api/#launch_nested_container_session) call.

When the `--tty` option is passed, the container is launched with a TTY. The local terminal is also set
in raw mode and restored before command termination. Lastly, the container will be launched with the `TERM`
env var set to `xterm`.

The `--user` option specifies the user to launch the nested container as.

The container ID is generated as a [version 4 UUID](https://en.wikipedia.org/wiki/Universally_unique_identifier#Version_4_(random))
and has as parent the container ID found in the aforementioned step.

The command blocks until the `LAUNCH_NESTED_CONTAINER_SESSION` call returns, and will only proceed with
further calls (eg. `ATTACH_CONTAINER_INPUT`) once the response headers have been received (which means
the nested container has been created).

The `LAUNCH_NESTED_CONTAINER_SESSION` response payload contains 2 types of `PROCESS_IO` messages:

- `DATA` messages of type `STDOUT`.
- `DATA` messages of type `STDERR`.

The data is then continously written to the command STDOUT or STDERR.

### ATTACH_CONTAINER_INPUT

When `--interactive` is passed, the command attaches its STDIN through an
[ATTACH_CONTAINER_INPUT](http://mesos.apache.org/documentation/latest/operator-http-api/#attach_container_input) call.

The first message sent is of type `CONTAINER_ID` and contains the container ID of the nested task created in the above step.

Afterwards, the command will continuously send 2 types of `PROCESS_IO` messages:

- `DATA` messages of type `STDIN`, which contain the data being read from STDIN.
- `CONTROL` messages of type `HEARTBEAT` are being sent every 30 seconds in order to keep the connection alive.

Furthermore, when the `--tty` option is passed, the command will send an initial pair of `CONTROL`
messages of type `TTY_INFO` to force a redraw of the remote terminal. The first one being of size [0,0]
and the second one being the actual size of the local terminal. `SIGWINCH` signals will also be caught
and will lead to appropriate `TTY_INFO` messages in order to redraw the remote terminal.

Still when the `--tty` option is passed, data being read from STDIN will be buffered in order to detect
a special `escape sequence` (`Ctrl+P`, `Ctrl+Q`). When encountered, all connections are closed and the
command exits with an exit code of `0`.

### Waiting for container exit status

Unless the calls to `ATTACH_CONTAINER_INPUT` or `LAUNCH_NESTED_CONTAINER_SESSION` were terminated because of
a termination signal or the escape sequence being detected, the command will make a `WAIT_CONTAINER` call
in order to determine an exit code.

The `WAIT_CONTAINER` returns an `exit_status`, which is the return value of `wait(2)`.

If the container was terminated because of a signal, the command exits with an exit code of `128 + signal_value`.

If the container terminated normally (that is, by calling exit(3), _exit(2), or by returning from main()),
the command exits with the same exit code.
