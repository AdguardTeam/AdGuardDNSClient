# AdGuard DNS Client changelog

All notable changes to this project will be documented in this file.

The format is based on [*Keep a Changelog*](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

<!--
## [v0.0.2] - 2024-06-29 (APPROX.)

See also the [v0.0.2 GitHub milestone][ms-v0.0.2].

[ms-v0.0.2]: https://github.com/AdguardTeam/AdGuardDNSClient/milestone/1?closed=1

NOTE: Add new changes BELOW THIS COMMENT.
-->

### Security

- Go version has been updated to prevent the possibility of exploiting the Go vulnerabilities fixed in [Go 1.22.5][go-1.22.5].

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

[go-1.22.5]: https://groups.google.com/g/golang-announce/c/gyb7aM1C9H4
[go-59229]:  https://github.com/golang/go/issues/59229

<!--
NOTE: Add new changes ABOVE THIS COMMENT.
-->

## [v0.0.1] - 2024-06-17

### Added

- Everything!

<!--
[Unreleased]: https://github.com/AdguardTeam/AdGuardDNSClient/compare/v0.0.2...HEAD
[v0.0.2]:     https://github.com/AdguardTeam/AdGuardDNSClient/compare/v0.0.1...v0.0.2
-->

[Unreleased]: https://github.com/AdguardTeam/AdGuardDNSClient/compare/v0.0.1...HEAD
[v0.0.1]:     https://github.com/AdguardTeam/AdGuardDNSClient/compare/v0.0.0...v0.0.1
