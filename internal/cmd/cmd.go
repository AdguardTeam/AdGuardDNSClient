// Package cmd is the AdGuardDNSClient entry point.
package cmd

import (
	"github.com/AdguardTeam/AdGuardDNSClient/internal/version"
	"github.com/AdguardTeam/golibs/log"
)

// Main is the entrypoint of AdGuardDNSClient.  Main may accept arguments, such as
// embedded assets and command-line arguments.
func Main() {
	// TODO(e.burkov): Use [log/slog] when [dnsproxy] will.
	log.Info(
		"AdGuardDNSClient version %q built from commit %q on branch %q at %q, race: %t",
		version.Version(),
		version.Revision(),
		version.Branch(),
		version.CommitTime(),
		version.RaceEnabled,
	)

	// TODO(e.burkov):  Actually use [check].
	check(nil)
}

// check is a simple error-checking helper.  It must only be used within Main.
func check(err error) {
	if err != nil {
		panic(err)
	}
}
