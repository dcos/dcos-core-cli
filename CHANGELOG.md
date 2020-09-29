# CHANGELOG

## 2.2-patch.x

* Breaking changes

    * `dcos service shutdown` asks for confirmation unless `--yes` flag specified (#506).

## 2.1-patch.1

* Breaking changes

    * Change default SSH user from `core` to `centos` (#448).

* Fixes

    * Update dependencies (#459, #478).
    * Update `dcos node diagnostics` deprecation message (#452).
    * Allow changing SSH user without using a proxy (#448).


## 2.1-patch.0

* Features

  * Added `dcos calico` command

* Notes

During this new minor release we also refactored the `dcos describe` subcommand from Python to Go.


## 2.0-patch.1

* Fixes

  * Support Docker `forcePullImage` in job JSON definitions
  * `dcos job add` and `dcos job update` should show the help menu by default

## 2.0-patch.0

* Deprecations

  * Deprecated --mesos-id in some commands

* Features

  * Added `dcos diagnostics` command
  * Added `SCARCE` column to `marathon debug details` command output (#341)
  * Support custom escape sequences (#331)
  * Expose task roles in Marathon subcommands.
  * Support custom escape sequences for `dcos task attach`

## 1.14-patch.4

* Fixes

  * Enforce role by default when creating a Marathon group
  * Improve error messages on metronome API errors
  * Fix error when detecting a partial escape sequence

## 1.14-patch.3

* Fixes

  * Make sure to consume remaining output when input connection fails during 'dcos task exec -i'

## 1.14-patch.2

* Fixes

  * Update 'dcos quota create' to error out less often according to internal UX feedback

## 1.14-patch.1

* Fixes

  * Improve the `dcos quota` subcommand according to internal UX feedback

## 1.14-patch.0

* Breaking changes

  * `dcos task ls` without any argument to get the list of all tasks files is not supported anymore

* Features

  * Introduce the `dcos quota` subcommand to manage DC/OS quotas
  * Add `dcos node drain` subcommand to drain nodes of their tasks
  * Add `dcos node reactivate` and `dcos node deactivate` to maintain agents
  * Add `--agent-id` to `dcos task list` to only list tasks on a given agent
  * Add `dcos task download` to download task sandbox files
  * Add a `--user` option to `dcos task exec` to specify the user for the launched nested container
  * Add an `--all` option to `dcos node log` to print all the log lines
  * Add job task ID(s) when printing history with `dcos job history --json`
  * Support `--id` when creating a group through `dcos marathon group add`

* Notes

During this new minor release we also refactored the `dcos task` and `dcos service` subcommands from Python to Go.

The only remaining subcommands in Python are `dcos marathon` and `dcos package`.
