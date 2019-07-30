# CHANGELOG

## 1.14-patch.0

* Breaking changes

  * `dcos task ls` without any argument to get the list of all tasks files is not supported anymore.

* Features

  * Introduce the `dcos quota` subcommand to manage Mesos quotas
  * Add `dcos node drain` subcommand to drain nodes of their tasks
  * Add `dcos node reactivate` and `dcos node deactivate` to ease maintenance
  * Add `--agent-id` to `dcos task list` to only list task on a given agent
  * Add `dcos task download` to download task sandbox files
  * Add a `--user` option to `dcos task exec` to specify the user for the launched nested container
  * Add an `--all` option to `dcos node log` to print all the log lines
  * Add job task ID(s) when printing history with `dcos job history --json`
  * Support --id when creating a group through `dcos marathon group add`

* Notes

During this new minor release we also refactored the `dcos task` and `dcos service` subcommands from Python to Go.

The only remaining subcommands in Python are `dcos marathon` and `dcos package`.
