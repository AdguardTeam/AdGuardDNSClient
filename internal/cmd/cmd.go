// Package cmd is the AdGuardDNSClient entry point.
package cmd

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
	"github.com/AdguardTeam/golibs/osutil"
	osservice "github.com/kardianos/service"
)

// Main is the entrypoint of AdGuardDNSClient.  Main may accept arguments, such
// as embedded assets and command-line arguments.
func Main() {
	// TODO(a.garipov):  Use for start cancelation.
	ctx := context.Background()

	opts, err := parseOptions()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		flag.CommandLine.SetOutput(os.Stderr)
		flag.CommandLine.Usage()

		os.Exit(statusArgumentError)
	}

	if opts.help {
		flag.CommandLine.SetOutput(os.Stdout)
		flag.CommandLine.Usage()

		os.Exit(osutil.ExitCodeSuccess)
	}

	envs, envsErrs := parseLogEnvs()
	l, logFile, envsLoggerErr := newEnvLogger(opts, envs)

	conf, err := handleServiceConfig(opts.serviceAction)
	l, logFile, confLoggerErrs := newConfigLogger(l, logFile, opts, envs, conf)

	prog := &program{
		conf:    conf,
		done:    make(chan struct{}),
		errCh:   make(chan error),
		log:     l,
		logFile: logFile,
	}

	reportPrevErrs(ctx, l, envsErrs, envsLoggerErr, confLoggerErrs)

	check(ctx, prog, err)

	svc, err := osservice.New(prog, newServiceConfig())
	check(ctx, prog, err)

	if opts.serviceAction != "" {
		exitCode := control(svc, opts.serviceAction)
		prog.closeLogs()

		os.Exit(exitCode)
	}

	err = svc.Run()
	check(ctx, prog, err)

	prog.closeLogs()
}

// reportPrevErrs reports errors that were collected while there was no logger.
func reportPrevErrs(
	ctx context.Context,
	l *slog.Logger,
	envsErrs []error,
	loggerErr error,
	confLoggerErrs []error,
) {
	if len(envsErrs) > 0 {
		l.ErrorContext(ctx, "parsing environment", slogutil.KeyError, errors.Join(envsErrs...))
	}

	if loggerErr != nil {
		l.ErrorContext(ctx, "creating env logger", slogutil.KeyError, loggerErr)
	}

	if len(confLoggerErrs) > 0 {
		l.ErrorContext(ctx, "creating conf logger", slogutil.KeyError, errors.Join(confLoggerErrs...))
	}
}

// check writes the error to the program's log and exits the process with a
// failure code if the error is not nil.  It also closes log files.  It must
// only be called within [Main].
func check(ctx context.Context, prog *program, err error) {
	if err == nil {
		return
	}

	prog.log.ErrorContext(ctx, "fatal error", slogutil.KeyError, err)
	prog.closeLogs()

	os.Exit(osutil.ExitCodeFailure)
}
