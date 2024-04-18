package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
	"github.com/AdguardTeam/golibs/service"
)

// serviceHandler shuts services down when the done channel is closed.
type serviceHandler struct {
	done            <-chan struct{}
	services        []service.Interface
	shutdownTimeout time.Duration
}

// serviceHandlerPrefix is the default and recommended prefix for the logger of
// a [serviceHandler].
const serviceHandlerPrefix = "service_handler"

// newServiceHandler returns a new properly initialized *serviceHandler that
// shuts down services.  The signal for shutting down is the close of done
// channel, it must not be nil.  timeout is the maximum time to wait for the
// services to shut down, avoid using 0.
func newServiceHandler(done <-chan struct{}, timeout time.Duration) (h *serviceHandler) {
	return &serviceHandler{
		done:            done,
		services:        nil,
		shutdownTimeout: timeout,
	}
}

// add adds a services to the signal handler.
//
// It must not be called concurrently with [serviceHandler.handle].
func (h *serviceHandler) add(svcs ...service.Interface) {
	h.services = append(h.services, svcs...)
}

// handle blocks until the termination channel is closed, after which it shuts
// down all services.  ctx is used for logging and serves as the base for the
// shutdown timeout.
//
// handle must not be called concurrently with [serviceHandler.add].
func (h *serviceHandler) handle(ctx context.Context, l *slog.Logger, errCh chan<- error) {
	defer slogutil.RecoverAndLog(ctx, l)

	if _, ok := <-h.done; ok {
		// Shouldn't happen, since h.done is currently only closed.
		panic("unexpected write to done channel")
	}

	l.InfoContext(ctx, "received shutdown signal")

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, h.shutdownTimeout)
	defer cancel()

	errCh <- h.shutdown(ctx, l)
}

// shutdown gracefully shuts down all services and returns all the occurred
// errors joined.
func (h *serviceHandler) shutdown(ctx context.Context, l *slog.Logger) (err error) {
	l.InfoContext(ctx, "shutting down")

	var errs []error
	for i := len(h.services) - 1; i >= 0; i-- {
		s := h.services[i]
		err = s.Shutdown(ctx)
		if err != nil {
			err = fmt.Errorf("service at index %d: %w", i, err)
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
