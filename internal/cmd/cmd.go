// Package cmd is the AdGuardDNSClient entry point.
package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/dnssvc"
	"github.com/AdguardTeam/AdGuardDNSClient/internal/version"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/log"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
	"github.com/AdguardTeam/golibs/service"
)

// logFormat is the current implementation of the logger.
//
// TODO(e.burkov):  Use [log/slog] in [dnsproxy] and remove this.
const logFormat = slogutil.FormatAdGuardLegacy

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

	// Logger

	// Error is always nil for the moment.
	logFmt, _ := slogutil.NewFormat(logFormat)

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

	// DNS service

	dnsSvc, err := dnssvc.New(conf.DNS.toInternal())
	check(err)

	err = dnsSvc.Start(ctx)
	check(err)
	l.DebugContext(ctx, "dns service started")

	// Signal handler
	//
	// TODO(e.burkov):  Add when it will support Windows.

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	// Wait for a signal.
	sig := <-sigCh
	l.DebugContext(ctx, "received signal", "value", sig)

	shutdown(ctx, l, dnsSvc)
}

// shutdown gracefully stops the services.
//
// TODO(e.burkov):  Use [service.SignalHandler] when it will support Windows.
func shutdown(ctx context.Context, l *slog.Logger, svcs ...service.Interface) {
	var errs []error
	for i, svc := range svcs {
		if svc == nil {
			continue
		}

		if err := svc.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("service at index %d: %w", i, err))
		}
	}

	if err := errors.Join(errs...); err != nil {
		l.ErrorContext(
			ctx,
			"shutting down",
			"error", err,
		)
	}
}
