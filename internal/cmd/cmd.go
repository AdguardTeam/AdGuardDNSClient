// Package cmd is the AdGuardDNSClient entry point.
package cmd

import (
	"context"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/version"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
)

// logFormat is the used implementation of the log.
//
// TODO(e.burkov):  Consider making configurable.
const logFormat = slogutil.FormatAdGuardLegacy

// Main is the entrypoint of AdGuardDNSClient.  Main may accept arguments, such as
// embedded assets and command-line arguments.
func Main() {
	ctx := context.Background()

	conf, err := parseConfiguration(defaultConfigPath)
	check(err)

	// Error is always nil for the moment.
	logFmt, _ := slogutil.NewFormat(logFormat)

	l := slogutil.New(&slogutil.Config{
		Format: logFmt,
		// TODO(e.burkov):  Configure timestamp.
		Verbose: conf.Log.Verbose,
	})

	// TODO(a.garipov): Copy logs configuration from the WIP abt. slog.
	buildVersion, revision, branch := version.Version(), version.Revision(), version.Branch()
	l.InfoContext(
		ctx,
		"go-proj-skel starting",
		"version", buildVersion,
		"revision", revision,
		"branch", branch,
		"commit_time", version.CommitTime(),
		"race", version.RaceEnabled,
	)
}

// check is a simple error-checking helper.  It must only be used within Main.
func check(err error) {
	if err != nil {
		panic(err)
	}
}
