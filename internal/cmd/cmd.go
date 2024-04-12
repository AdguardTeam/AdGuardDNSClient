// Package cmd is the AdGuardDNSClient entry point.
package cmd

import (
	"context"
	"os"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/dnssvc"
	"github.com/AdguardTeam/AdGuardDNSClient/internal/version"
	"github.com/AdguardTeam/golibs/log"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
	"github.com/AdguardTeam/golibs/service"
)

// Main is the entrypoint of AdGuardDNSClient.  Main may accept arguments, such
// as embedded assets and command-line arguments.
func Main() {
	opts, err := parseOptions()
	exitCode, needsExit := processOptions(opts, err)
	if needsExit {
		os.Exit(exitCode)
	}

	conf, err := parseConfig(defaultConfigPath)
	check(err)

	err = conf.validate()
	check(err)

	// Error is always nil for the moment.
	//
	// TODO(e.burkov):  Use [log/slog] in [dnsproxy] and then change it.
	logFmt, _ := slogutil.NewFormat(slogutil.FormatAdGuardLegacy)

	// TODO(e.burkov):  Configure timestamp and output.
	isVerbose := opts.verbose || conf.Log.Verbose
	l := slogutil.New(&slogutil.Config{
		Format:  logFmt,
		Verbose: isVerbose,
	})
	if isVerbose {
		log.SetLevel(log.DEBUG)
	}

	ctx := context.Background()

	// TODO(a.garipov): Copy logs configuration from the WIP abt. slog.
	l.InfoContext(
		ctx,
		"AdGuardDNSClient starting",
		"version", version.Version(),
		"revision", version.Revision(),
		"branch", version.Branch(),
		"commit_time", version.CommitTime(),
		"race", version.RaceEnabled,
		"verbose", isVerbose,
	)

	sigHdlr := service.NewSignalHandler(&service.SignalHandlerConfig{
		Logger: l.With(slogutil.KeyPrefix, service.SignalHandlerPrefix),
	})

	dnsSvc, err := dnssvc.New(conf.DNS.toInternal())
	check(err)

	err = dnsSvc.Start(ctx)
	check(err)

	sigHdlr.Add(dnsSvc)
	l.DebugContext(ctx, "dns service started")

	os.Exit(sigHdlr.Handle(ctx))
}
