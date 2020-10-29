# CHANGELOG

## 2.0-patch.x

* Fixes

  * Fix NaN% in quota list (#519)

## 2.0-patch.8

* Fixes

    * Update `dcos node diagnostics` deprecation message
    * Display version for packages with the same name

## 2.0-patch.7

* Fixes

    * Allow changing SSH user without using a proxy
    * Update PyInstaller (CVE-2019-16784)

## 2.0-patch.6

* Fixes

    * Fix "cannot unmarshal number" for packages that use a unix timestamp as their `releaseVersion`

## 2.0-patch.5

* Fixes

  * Make local installed apps ovverride Cosmos Apps (#405)

## 2.0-patch.4

* Features

  * Add a `CERTIFIED` column to `dcos package list` and `dcos package search` commands

* Fixes

  * Add better error descriptions for `node diagnostics` command.
  * Fix the `--app-id` filter in `dcos package list`
  * Fixing incorrect DC/OS version string in diagnostics subcommand

## 2.0-patch.3

* Fixes

  * Add option to create master/agents only diagnostics bundle #372
  * Fixed creation of nested groups, the regression was introduced in #322
  * Support more fields in job JSON definition.

## 2.0-patch.2

This release reverts https://github.com/dcos/dcos-core-cli/pull/354.
The commit has landed in the 2.1-patch.x branch instead.

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
