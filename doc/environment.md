# AdGuard DNS Client environment configuration

AdGuard DNS Client uses [environment variables][wiki-env] to store some of the configuration. All other configuration is stored in the [configuration file][conf].

## Contents

- [`LOG_OUTPUT`](#LOG_OUTPUT)
- [`LOG_FORMAT`](#LOG_FORMAT)
- [`LOG_TIMESTAMP`](#LOG_TIMESTAMP)
- [`VERBOSE`](#VERBOSE)

[conf]:     configuration.md
[wiki-env]: https://en.wikipedia.org/wiki/Environment_variable

> [!NOTE]
> In its current state, the log system is only intended for debugging startup errors.

## <a href="#LOG_OUTPUT" id="LOG_OUTPUT" name="LOG_OUTPUT">`LOG_OUTPUT`</a>

The log destination, must be an absolute path to the file or one of the special values.

- `syslog` means that the platform-specific system log is used, which is syslog for Linux and Event Log for Windows.

    > [!NOTE]
    > Log entries written to the system log are in text format and use the system timestamp.

- `stdout` for standard output stream.

- `stderr` for standard error stream.

- Absolute path to the log file.

    **Example:** `/home/user/logs`.

    **Example:** `C:\Users\user\logs.txt`.

This environment variable has priority over [log.output][conf-log-output] field from the configuration file.

**Default:** **Unset.**

[conf-log-output]: configuration.md#log-output

## <a href="#LOG_FORMAT" id="LOG_FORMAT" name="LOG_FORMAT">`LOG_FORMAT`</a>

The format for log entries.

- `adguard_legacy`;
- `default`;
- `json`;
- `json_hybrid`;
- `text`.

<!--
    TODO(s.chzhen):  Add output examples.
-->

This environment variable has priority over [log.format][conf-log-format] field from the configuration file.

**Default:** **Unset.**

[conf-log-format]: configuration.md#log-format

## <a href="#LOG_TIMESTAMP" id="LOG_TIMESTAMP" name="LOG_TIMESTAMP">`LOG_TIMESTAMP`</a>

When set to `1`, log entries have a timestamp.  When set to `0`, log entries don’t have it.

This environment variable has priority over [log.timestamp][conf-log-timestamp] field from the configuration file.

**Default:** **Unset.**

[conf-log-timestamp]: configuration.md#log-timestamp

## <a href="#VERBOSE" id="VERBOSE" name="VERBOSE">`VERBOSE`</a>

When set to `1`, enable verbose logging.  When set to `0`, disable it.

This environment variable has priority over [log.verbose][conf-log-verbose] field from the configuration file.

**Default:** **Unset.**

[conf-log-verbose]: configuration.md#log-verbose
