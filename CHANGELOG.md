# AdGuard DNS Client changelog

All notable changes to this project will be documented in this file.

The format is based on [*Keep a Changelog*](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

<!--
## [v0.0.4] - 2025-04-30 (APPROX.)

See also the [v0.0.4 GitHub milestone][ms-v0.0.4].

[ms-v0.0.4]: https://github.com/AdguardTeam/AdGuardDNSClient/milestone/4?closed=1

NOTE: Add new changes BELOW THIS COMMENT.
-->

<!--
NOTE: Add new changes ABOVE THIS COMMENT.
-->

## [v0.0.3] - 2025-03-31

See also the [v0.0.3 GitHub milestone][ms-v0.0.3].

### Security

- Go version has been updated to prevent the possibility of exploiting the Go vulnerabilities fixed in [Go 1.24.1][go-1.24.1].

### Changed

#### Configuration changes

In this release, the schema version has changed from 1 to 2.

- The new object `bind_retry` has been added to the `dns.server` object.

    ```yaml
    # BEFORE:
    dns:
        server:
            # …
        # …
    # …
    schema_version: 1

    # AFTER:
    dns:
        server:
            bind_retry:
                enabled: true
                interval: 1s
                count: 4
            # …
        # …
    # …
    schema_version: 2
    ```

To rollback this change, remove the `dns.server.bind_retry` object and set the `schema_version` to `1`.

### Fixed

- Failed binding to listen addresses when installed as Windows service ([#11]).

[#11]: https://github.com/AdguardTeam/AdGuardDNSClient/issues/11

[go-1.24.1]: https://groups.google.com/g/golang-announce/c/4t3lzH3I0eI
[ms-v0.0.3]: https://github.com/AdguardTeam/AdGuardDNSClient/milestone/3?closed=1

## [v0.0.2] - 2024-11-08

See also the [v0.0.2 GitHub milestone][ms-v0.0.2].

### Security

- Go version has been updated to prevent the possibility of exploiting the Go vulnerabilities fixed in [Go 1.23.3][go-1.23.3].

### Added

- MSI installer for the ARM64 architecture in addition to the existing x86 and x64 installers.

### Changed

- Path to the executable is now validated when the application installs itself as a `launchd` service on macOS ([#2]).

### Fixed

- The `syslog` log output on macOS ([#3]).

  **NOTE:** The implementation is actually a workaround for a known [Go issue][go-59229], and uses the `/usr/bin/logger` utility. This approach is suboptimal and will be improved once the Go issue is resolved.
- DNS proxy logs being written to `stderr` instead of `log.output` ([#1]).

[#1]: https://github.com/AdguardTeam/AdGuardDNSClient/issues/1
[#2]: https://github.com/AdguardTeam/AdGuardDNSClient/issues/2
[#3]: https://github.com/AdguardTeam/AdGuardDNSClient/issues/3

[go-1.23.3]: https://groups.google.com/g/golang-announce/c/X5KodEJYuqI
[go-59229]:  https://github.com/golang/go/issues/59229
[ms-v0.0.2]: https://github.com/AdguardTeam/AdGuardDNSClient/milestone/1?closed=1

## [v0.0.1] - 2024-06-17

### Added

- Everything!

<!--
[Unreleased]: https://github.com/AdguardTeam/AdGuardDNSClient/compare/v0.0.4...HEAD
[v0.0.4]:     https://github.com/AdguardTeam/AdGuardDNSClient/compare/v0.0.3...v0.0.4
-->

[Unreleased]: https://github.com/AdguardTeam/AdGuardDNSClient/compare/v0.0.3...HEAD
[v0.0.3]:     https://github.com/AdguardTeam/AdGuardDNSClient/compare/v0.0.2...v0.0.3
[v0.0.2]:     https://github.com/AdguardTeam/AdGuardDNSClient/compare/v0.0.1...v0.0.2
[v0.0.1]:     https://github.com/AdguardTeam/AdGuardDNSClient/compare/v0.0.0...v0.0.1
