package cmd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/dnssvc"
	"github.com/AdguardTeam/AdGuardDNSClient/internal/version"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
	"github.com/AdguardTeam/golibs/service"
	osservice "github.com/kardianos/service"
)

// program is the implementation of the [osservice.Interface] interface for
// AdGuardDNSClient.
type program struct {
	// TODO(e.burkov):  Add *options?

	// conf is the parsed configuration to run the program.  It appears nil on
	// any service action and must not be accessed.
	conf    *configuration
	logger  *slog.Logger
	logFile *os.File
	done    chan struct{}
	errCh   chan error
}

// type check
var _ osservice.Interface = (*program)(nil)

// serviceProgramPrefix is the default and recommended prefix for the logger of
// the default [osservice.Interface] implementation.
const serviceProgramPrefix = "program"

// Start implements the [osservice.Interface] interface for [*program].
func (prog *program) Start(_ osservice.Service) (err error) {
	ctx := context.Background()
	l := prog.logger.With(slogutil.KeyPrefix, serviceProgramPrefix)

	// TODO(a.garipov): Copy logs configuration from the WIP abt. slog.
	l.InfoContext(
		ctx,
		"AdGuardDNSClient starting",
		"version", version.Version(),
		"revision", version.Revision(),
		"branch", version.Branch(),
		"commit_time", version.CommitTime(),
		"race", version.RaceEnabled,
		"verbose", l.Enabled(ctx, slog.LevelDebug),
	)

	svcHdlr := newServiceHandler(prog.done, service.SignalHandlerShutdownTimeout)

	dnsSvc, err := dnssvc.New(prog.conf.DNS.toInternal(prog.logger))
	if err != nil {
		return fmt.Errorf("creating dns service: %w", err)
	}

	err = dnsSvc.Start(ctx)
	if err != nil {
		return fmt.Errorf("starting dns service: %w", err)
	}

	svcHdlr.add(dnsSvc)
	l.DebugContext(ctx, "dns service started")

	go svcHdlr.handle(ctx, prog.logger.With(slogutil.KeyPrefix, "service_handler"), prog.errCh)

	return nil
}

// Stop implements the [osservice.Interface] interface for [*program].
func (prog *program) Stop(_ osservice.Service) (err error) {
	close(prog.done)

	return <-prog.errCh
}

// closeLogs closes the log files and syslog handler, if there are any.
func (prog *program) closeLogs() {
	// At this point, just use stderr with defaults.
	l := slogutil.New(&slogutil.Config{
		Output: os.Stderr,
	}).With(slogutil.KeyPrefix, serviceProgramPrefix)

	if prog.logFile != nil {
		err := prog.logFile.Close()
		if err != nil {
			l.Error("stopping: closing log file", slogutil.KeyError, err)
		}
	}

	h := prog.logger.Handler()
	if c, ok := h.(io.Closer); ok {
		err := c.Close()
		if err != nil {
			l.Error("stopping: closing syslog", slogutil.KeyError, err)
		}
	}
}
