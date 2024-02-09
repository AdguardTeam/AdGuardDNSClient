// Package cmd is the AdGuardDNSClient entry point.
package cmd

import (
	"context"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/version"
	"github.com/AdguardTeam/golibs/log"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
)

// logFormat is the used implementation of the log.
//
// TODO(e.burkov):  Use [log/slog] in [dnsproxy] and remove this.
const logFormat = slogutil.FormatAdGuardLegacy

// Main is the entrypoint of AdGuardDNSClient.  Main may accept arguments, such
// as embedded assets and command-line arguments.
func Main() {
	ctx := context.Background()

	conf, err := parseConfig(defaultConfigPath)
	check(err)

	check(conf.validate())

	// Error is always nil for the moment.
	logFmt, _ := slogutil.NewFormat(logFormat)

	// TODO(e.burkov):  Configure timestamp and output.
	l := slogutil.New(&slogutil.Config{
		Format:  logFmt,
		Verbose: conf.Log.Verbose,
	})
	if conf.Log.Verbose {
		log.SetLevel(log.DEBUG)
	}

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
		"verbose", conf.Log.Verbose,
	)
}
