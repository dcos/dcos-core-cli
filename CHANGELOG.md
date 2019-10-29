# CHANGELOG

## 1.13-patch.x

* Fixes

  * Add better error descriptions for `node diagnostics` command.
  * Support more fields in job JSON definition.

## 1.13-patch.6

* Fixes

  * Support Docker `forcePullImage` in job JSON definitions
  * `dcos job add` and `dcos job update` should show the help menu by default

## 1.13-patch.5

* Fixes

  * Fix possible segfault in `node ssh`
  * Fix unknown subcommand errors
  * Upgrade to Python 3.7 on UNIX systems
  * Updates get_package_commands to check plugin.toml
  * Uses job's task ID(s) when printing history with json
  * Support SSE heartbeats

## 1.13-patch.4

* Fixes

  * Print the version mismatch warning to stderr.

## 1.13-patch.3

* Fixes

  * Add dynamic autocompletion to dcos job
  * Added 'GPUs' to Mesos Resources struct

## 1.13-patch.2

* Fixes

  * Fix `volumes` field and add `ucr` field in job JSON.
  * Revert "Specified if agent is public when using 'dcos node list'"

## 1.13-patch.1

* Fixes

  * `dcos node list` shouldn't error-out if it misses public IPs
  * Fixed _log() to have old behavior when dcos-log disabled.

## 1.13-patch.0

* Features

  * Add color support to `dcos node log`
  * Add a public IP field to `dcos node list`
  * Add `--user` flag to `dcos service log`
  * Add journalctl format options to `dcos node log`: `json-pretty`, `json`, `cat`, `short`
